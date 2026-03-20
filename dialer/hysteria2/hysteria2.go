package hysteria2

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mzz2017/gg/common"
	"github.com/mzz2017/gg/dialer"
	ggs "github.com/mzz2017/gg/dialer/sagernet"
	shy2 "github.com/sagernet/sing-quic/hysteria2"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func init() {
	dialer.FromLinkRegister("hysteria2", NewHysteria2)
	dialer.FromLinkRegister("hy2", NewHysteria2)
}

type Hysteria2 struct {
	Name          string        `json:"name"`
	Server        string        `json:"server"`
	Port          int           `json:"port"`
	ServerPorts   []string      `json:"serverPorts"`
	Password      string        `json:"password"`
	SNI           string        `json:"sni"`
	ALPN          []string      `json:"alpn,omitempty"`
	Obfs          string        `json:"obfs,omitempty"`
	ObfsPassword  string        `json:"obfsPassword,omitempty"`
	PinSHA256     string        `json:"pinSHA256,omitempty"`
	AllowInsecure bool          `json:"allowInsecure"`
	UpMbps        int           `json:"upMbps,omitempty"`
	DownMbps      int           `json:"downMbps,omitempty"`
	HopInterval   time.Duration `json:"hopInterval,omitempty"`
	Protocol      string        `json:"protocol"`
}

func NewHysteria2(link string, opt *dialer.GlobalOption) (*dialer.Dialer, error) {
	s, err := ParseHysteria2URL(link)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", dialer.InvalidParameterErr, err)
	}
	if opt.AllowInsecure {
		s.AllowInsecure = true
	}
	return s.Dialer()
}

func ParseHysteria2URL(link string) (*Hysteria2, error) {
	schemeEnd := strings.Index(link, "://")
	if schemeEnd == -1 {
		return nil, dialer.InvalidParameterErr
	}
	scheme := link[:schemeEnd]
	if scheme != "hysteria2" && scheme != "hy2" {
		return nil, dialer.InvalidParameterErr
	}
	rawAfterScheme := link[schemeEnd+3:]
	authorityEnd := strings.IndexAny(rawAfterScheme, "/?#")
	rawAuthority := rawAfterScheme
	rawSuffix := ""
	if authorityEnd != -1 {
		rawAuthority = rawAfterScheme[:authorityEnd]
		rawSuffix = rawAfterScheme[authorityEnd:]
	}
	userHost := rawAuthority
	if at := strings.LastIndex(rawAuthority, "@"); at != -1 {
		userHost = rawAuthority[at+1:]
	}
	host, portSpec, err := splitHostPortSpec(userHost)
	if err != nil {
		return nil, err
	}
	sanitizedAuthority := rawAuthority
	if comma := strings.Index(portSpec, ","); comma != -1 {
		sanitizedAuthority = strings.TrimSuffix(rawAuthority, portSpec) + portSpec[:comma]
	}
	u, err := url.Parse(scheme + "://" + sanitizedAuthority + rawSuffix)
	if err != nil {
		return nil, err
	}
	port, serverPorts, err := parsePortSpec(portSpec)
	if err != nil {
		return nil, err
	}
	password := ""
	if u.User != nil {
		if p, ok := u.User.Password(); ok {
			password = u.User.Username() + ":" + p
		} else {
			password = u.User.Username()
		}
	}
	if password == "" {
		return nil, fmt.Errorf("missing auth password")
	}
	hopInterval := time.Duration(0)
	if interval := u.Query().Get("hop-interval"); interval != "" {
		hopInterval, err = time.ParseDuration(interval)
		if err != nil {
			return nil, fmt.Errorf("invalid hop-interval: %w", err)
		}
	}
	return &Hysteria2{
		Name:          u.Fragment,
		Server:        host,
		Port:          port,
		ServerPorts:   serverPorts,
		Password:      password,
		SNI:           u.Query().Get("sni"),
		ALPN:          ggs.SplitCSV(u.Query().Get("alpn")),
		Obfs:          u.Query().Get("obfs"),
		ObfsPassword:  u.Query().Get("obfs-password"),
		PinSHA256:     u.Query().Get("pinSHA256"),
		AllowInsecure: common.StringToBool(u.Query().Get("insecure")) || common.StringToBool(u.Query().Get("allowInsecure")),
		UpMbps:        parseIntDefault(u.Query().Get("upmbps"), 0),
		DownMbps:      parseIntDefault(u.Query().Get("downmbps"), 0),
		HopInterval:   hopInterval,
		Protocol:      "hysteria2",
	}, nil
}

func splitHostPortSpec(hostport string) (string, string, error) {
	if hostport == "" {
		return "", "", fmt.Errorf("missing host")
	}
	if strings.HasPrefix(hostport, "[") {
		end := strings.Index(hostport, "]")
		if end == -1 {
			return "", "", fmt.Errorf("invalid IPv6 host")
		}
		host := hostport[1:end]
		portSpec := strings.TrimPrefix(hostport[end+1:], ":")
		return host, portSpec, nil
	}
	lastColon := strings.LastIndex(hostport, ":")
	if lastColon == -1 {
		return hostport, "", nil
	}
	return hostport[:lastColon], hostport[lastColon+1:], nil
}

func parsePortSpec(portSpec string) (int, []string, error) {
	if portSpec == "" {
		return 443, nil, nil
	}
	parts := strings.Split(portSpec, ",")
	first, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, nil, fmt.Errorf("invalid port: %w", err)
	}
	serverPorts := make([]string, 0, len(parts))
	if len(parts) > 1 {
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if part != "" {
				serverPorts = append(serverPorts, normalizePortRange(part))
			}
		}
	}
	return first, serverPorts, nil
}

func normalizePortRange(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "-") {
		return strings.Replace(value, "-", ":", 1)
	}
	if strings.Contains(value, ":") {
		return value
	}
	return value + ":" + value
}

func displayPortRange(value string) string {
	if strings.Contains(value, ":") {
		parts := strings.SplitN(value, ":", 2)
		if len(parts) == 2 && parts[0] == parts[1] {
			return parts[0]
		}
		return strings.Replace(value, ":", "-", 1)
	}
	return value
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (s *Hysteria2) ExportToURL() string {
	portSpec := strconv.Itoa(s.Port)
	if len(s.ServerPorts) > 0 {
		display := make([]string, 0, len(s.ServerPorts))
		for _, portRange := range s.ServerPorts {
			display = append(display, displayPortRange(portRange))
		}
		portSpec += "," + strings.Join(display, ",")
	}
	query := url.Values{}
	common.SetValue(&query, "sni", s.SNI)
	common.SetValue(&query, "alpn", strings.Join(s.ALPN, ","))
	common.SetValue(&query, "obfs", s.Obfs)
	common.SetValue(&query, "obfs-password", s.ObfsPassword)
	common.SetValue(&query, "pinSHA256", s.PinSHA256)
	common.SetValue(&query, "insecure", common.BoolToString(s.AllowInsecure))
	if s.UpMbps > 0 {
		common.SetValue(&query, "upmbps", strconv.Itoa(s.UpMbps))
	}
	if s.DownMbps > 0 {
		common.SetValue(&query, "downmbps", strconv.Itoa(s.DownMbps))
	}
	if s.HopInterval > 0 {
		common.SetValue(&query, "hop-interval", s.HopInterval.String())
	}
	return (&url.URL{
		Scheme:   "hysteria2",
		User:     url.User(s.Password),
		Host:     net.JoinHostPort(s.Server, portSpec),
		Path:     "/",
		RawQuery: query.Encode(),
		Fragment: s.Name,
	}).String()
}

func (s *Hysteria2) Dialer() (*dialer.Dialer, error) {
	serverName := s.SNI
	if serverName == "" {
		serverName = s.Server
	}
	tlsConfig, err := ggs.NewStdTLSConfig(serverName, s.AllowInsecure, s.ALPN, s.PinSHA256)
	if err != nil {
		return nil, err
	}
	client, err := shy2.NewClient(shy2.ClientOptions{
		Context:            context.Background(),
		Dialer:             N.SystemDialer,
		Logger:             ggs.Logger,
		ServerAddress:      M.ParseSocksaddrHostPort(s.Server, uint16(s.Port)),
		ServerPorts:        s.ServerPorts,
		HopInterval:        s.HopInterval,
		SendBPS:            uint64(s.UpMbps * 125000),
		ReceiveBPS:         uint64(s.DownMbps * 125000),
		SalamanderPassword: s.ObfsPassword,
		Password:           s.Password,
		TLSConfig:          tlsConfig,
	})
	if err != nil {
		return nil, err
	}
	return dialer.NewDialer(&clientDialer{
		client: client,
	}, true, s.Name, s.Protocol, s.ExportToURL()), nil
}

type clientDialer struct {
	client *shy2.Client
}

func (d *clientDialer) Dial(network, addr string) (net.Conn, error) {
	dst := M.ParseSocksaddr(addr)
	switch network {
	case "tcp":
		return d.client.DialConn(context.Background(), dst)
	case "udp":
		packetConn, err := d.client.ListenPacket(context.Background())
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(packetConn, dst.UDPAddr()), nil
	default:
		return nil, net.UnknownNetworkError(network)
	}
}

package anytls

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	sanytls "github.com/anytls/sing-anytls"
	"github.com/mzz2017/gg/common"
	"github.com/mzz2017/gg/dialer"
	ggs "github.com/mzz2017/gg/dialer/sagernet"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/uot"
)

func init() {
	dialer.FromLinkRegister("anytls", NewAnyTLS)
}

type AnyTLS struct {
	Name          string `json:"name"`
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Password      string `json:"password"`
	SNI           string `json:"sni"`
	AllowInsecure bool   `json:"allowInsecure"`
	Protocol      string `json:"protocol"`
}

func NewAnyTLS(link string, opt *dialer.GlobalOption) (*dialer.Dialer, error) {
	s, err := ParseAnyTLSURL(link)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", dialer.InvalidParameterErr, err)
	}
	if opt.AllowInsecure {
		s.AllowInsecure = true
	}
	return s.Dialer()
}

func ParseAnyTLSURL(link string) (*AnyTLS, error) {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "anytls" {
		return nil, fmt.Errorf("%w: %v", dialer.InvalidParameterErr, err)
	}
	port := 443
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("error when parsing port: %w", err)
		}
	}
	password := ""
	if u.User != nil {
		password = u.User.Username()
	}
	if password == "" {
		return nil, fmt.Errorf("missing auth password")
	}
	return &AnyTLS{
		Name:          u.Fragment,
		Server:        u.Hostname(),
		Port:          port,
		Password:      password,
		SNI:           u.Query().Get("sni"),
		AllowInsecure: common.StringToBool(u.Query().Get("insecure")) || common.StringToBool(u.Query().Get("allowInsecure")),
		Protocol:      "anytls",
	}, nil
}

func (s *AnyTLS) ExportToURL() string {
	query := url.Values{}
	common.SetValue(&query, "sni", s.SNI)
	common.SetValue(&query, "insecure", common.BoolToString(s.AllowInsecure))
	return (&url.URL{
		Scheme:   "anytls",
		User:     url.User(s.Password),
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Path:     "/",
		RawQuery: query.Encode(),
		Fragment: s.Name,
	}).String()
}

func (s *AnyTLS) Dialer() (*dialer.Dialer, error) {
	serverName := s.SNI
	if serverName == "" {
		serverName = s.Server
	}
	tlsConfig, err := ggs.NewStdTLSConfig(serverName, s.AllowInsecure, nil, "")
	if err != nil {
		return nil, err
	}
	server := M.ParseSocksaddrHostPort(s.Server, uint16(s.Port))
	client, err := sanytls.NewClient(context.Background(), sanytls.ClientConfig{
		Password: s.Password,
		DialOut: func(ctx context.Context) (net.Conn, error) {
			conn, err := N.SystemDialer.DialContext(ctx, N.NetworkTCP, server)
			if err != nil {
				return nil, err
			}
			tlsConn, err := ggs.ClientHandshake(ctx, conn, tlsConfig)
			if err != nil {
				_ = conn.Close()
				return nil, err
			}
			return tlsConn, nil
		},
		Logger: ggs.Logger,
	})
	if err != nil {
		return nil, err
	}
	uotClient := &uot.Client{
		Dialer:  anyTLSDialer(client.CreateProxy),
		Version: uot.Version,
	}
	return dialer.NewDialer(&clientDialer{
		name:     s.Name,
		link:     s.ExportToURL(),
		client:   client,
		uot:      uotClient,
		protocol: s.Protocol,
	}, true, s.Name, s.Protocol, s.ExportToURL()), nil
}

type anyTLSDialer func(ctx context.Context, destination M.Socksaddr) (net.Conn, error)

func (d anyTLSDialer) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	return d(ctx, destination)
}

func (d anyTLSDialer) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	return nil, fmt.Errorf("udp listen is unsupported directly")
}

type clientDialer struct {
	name     string
	link     string
	protocol string
	client   *sanytls.Client
	uot      *uot.Client
}

func (d *clientDialer) Dial(network, addr string) (net.Conn, error) {
	dst := M.ParseSocksaddr(addr)
	switch network {
	case "tcp":
		return d.client.CreateProxy(context.Background(), dst)
	case "udp":
		packetConn, err := d.uot.ListenPacket(context.Background(), dst)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(packetConn, dst.UDPAddr()), nil
	default:
		return nil, net.UnknownNetworkError(network)
	}
}

package v2ray

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mzz2017/gg/dialer"
	ggs "github.com/mzz2017/gg/dialer/sagernet"
	"github.com/sagernet/sing-box/adapter"
	sbtls "github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	sbv2ray "github.com/sagernet/sing-box/transport/v2ray"
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common/json/badoption"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type sagerNetV2RayDialer struct {
	server    M.Socksaddr
	tlsConfig sbtls.Config
	tlsDialer sbtls.Dialer
	transport adapter.V2RayClientTransport
	vmess     *vmess.Client
	vless     *vless.Client
}

func newSagerNetV2RayDialer(s *V2Ray) (*dialer.Dialer, error) {
	port, err := strconv.Atoi(s.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	server := M.ParseSocksaddrHostPort(s.Add, uint16(port))

	rd := &sagerNetV2RayDialer{
		server: server,
	}

	switch s.TLS {
	case "", "none", "tls", "xtls":
	default:
		return nil, fmt.Errorf("%w: security: %v", dialer.UnexpectedFieldErr, s.TLS)
	}

	needsTLS := needsV2RayTLS(s)
	if needsTLS {
		serverName := v2RayServerName(s)
		tlsConfig, err := ggs.NewStdTLSConfig(serverName, s.AllowInsecure, ggs.SplitCSV(s.Alpn), "")
		if err != nil {
			return nil, err
		}
		rd.tlsConfig = tlsConfig
		rd.tlsDialer = sbtls.NewDialer(N.SystemDialer, tlsConfig)
	}

	transport, err := newV2RayTransportOptions(s)
	if err != nil {
		return nil, err
	}
	return newSagerNetV2RayDialerWithTransport(s, rd, transport)
}

func newSagerNetV2RayDialerWithTransport(s *V2Ray, rd *sagerNetV2RayDialer, transport option.V2RayTransportOptions) (*dialer.Dialer, error) {
	if transport.Type != "" {
		var err error
		rd.transport, err = sbv2ray.NewClientTransport(context.Background(), N.SystemDialer, rd.server, transport, rd.tlsConfig)
		if err != nil {
			return nil, err
		}
	}

	switch s.Protocol {
	case "vmess":
		alterID := 0
		if s.Aid != "" {
			parsedAlterID, err := strconv.Atoi(s.Aid)
			if err != nil {
				return nil, fmt.Errorf("invalid alterId: %w", err)
			}
			alterID = parsedAlterID
		}
		client, err := vmess.NewClient(s.ID, "aes-128-gcm", alterID)
		if err != nil {
			return nil, err
		}
		rd.vmess = client
	case "vless":
		client, err := vless.NewClient(s.ID, "", ggs.Logger)
		if err != nil {
			return nil, err
		}
		rd.vless = client
	default:
		return nil, fmt.Errorf("%w: protocol: %v", dialer.UnexpectedFieldErr, s.Protocol)
	}

	return dialer.NewDialer(rd, true, s.Ps, s.Protocol, s.ExportToURL()), nil
}

func (d *sagerNetV2RayDialer) Dial(network, addr string) (net.Conn, error) {
	conn, err := d.dialServer()
	if err != nil {
		return nil, err
	}
	destination := M.ParseSocksaddr(addr)
	switch {
	case d.vmess != nil:
		switch network {
		case "tcp":
			return d.vmess.DialEarlyConn(conn, destination), nil
		case "udp":
			return d.vmess.DialEarlyPacketConn(conn, destination), nil
		}
	case d.vless != nil:
		switch network {
		case "tcp":
			return d.vless.DialEarlyConn(conn, destination)
		case "udp":
			return d.vless.DialEarlyPacketConn(conn, destination)
		}
	}
	_ = conn.Close()
	return nil, net.UnknownNetworkError(network)
}

func (d *sagerNetV2RayDialer) dialServer() (net.Conn, error) {
	switch {
	case d.transport != nil:
		return d.transport.DialContext(context.Background())
	case d.tlsDialer != nil:
		return d.tlsDialer.DialTLSContext(context.Background(), d.server)
	default:
		return N.SystemDialer.DialContext(context.Background(), N.NetworkTCP, d.server)
	}
}

func needsV2RayTLS(s *V2Ray) bool {
	switch s.TLS {
	case "", "none":
	case "tls", "xtls":
		return true
	default:
		return false
	}
	switch strings.ToLower(s.Net) {
	case "grpc", "h2":
		return true
	default:
		return false
	}
}

func v2RayServerName(s *V2Ray) string {
	if s.SNI != "" {
		return s.SNI
	}
	if s.Host != "" {
		return s.Host
	}
	return s.Add
}

func newV2RayTransportOptions(s *V2Ray) (option.V2RayTransportOptions, error) {
	var transport option.V2RayTransportOptions
	switch strings.ToLower(s.Net) {
	case "", "tcp":
		if s.Type != "none" && s.Type != "" {
			return transport, fmt.Errorf("%w: type: %v", dialer.UnexpectedFieldErr, s.Type)
		}
	case "ws", "websocket":
		transport.Type = C.V2RayTransportTypeWebsocket
		transport.WebsocketOptions.Path = s.Path
		if s.Host != "" {
			transport.WebsocketOptions.Headers = badoption.HTTPHeader{
				"Host": badoption.Listable[string]{s.Host},
			}
		}
	case "grpc":
		transport.Type = C.V2RayTransportTypeGRPC
		if s.Path != "" {
			transport.GRPCOptions.ServiceName = s.Path
		} else {
			transport.GRPCOptions.ServiceName = "GunService"
		}
	case "http", "h2":
		transport.Type = C.V2RayTransportTypeHTTP
		transport.HTTPOptions.Path = s.Path
		if s.Host != "" {
			transport.HTTPOptions.Host = badoption.Listable[string](ggs.SplitCSV(s.Host))
		}
	default:
		return transport, fmt.Errorf("%w: network: %v", dialer.UnexpectedFieldErr, s.Net)
	}
	return transport, nil
}

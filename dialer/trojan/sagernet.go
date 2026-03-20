package trojan

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mzz2017/gg/dialer"
	ggs "github.com/mzz2017/gg/dialer/sagernet"
	"github.com/sagernet/sing-box/adapter"
	sbtls "github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	sbtrojan "github.com/sagernet/sing-box/transport/trojan"
	sbv2ray "github.com/sagernet/sing-box/transport/v2ray"
	ss "github.com/sagernet/sing-shadowsocks2"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/json/badoption"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type sagerNetTrojanDialer struct {
	server    M.Socksaddr
	key       [sbtrojan.KeyLength]byte
	tlsDialer sbtls.Dialer
	transport adapter.V2RayClientTransport
	method    ss.Method
}

func newSagerNetTrojanDialer(s *Trojan) (*dialer.Dialer, error) {
	server := M.ParseSocksaddrHostPort(s.Server, uint16(s.Port))
	serverName := s.Sni
	if serverName == "" {
		serverName = s.Server
	}
	tlsConfig, err := ggs.NewStdTLSConfig(serverName, s.AllowInsecure, nil, "")
	if err != nil {
		return nil, err
	}

	rd := &sagerNetTrojanDialer{
		server:    server,
		key:       sbtrojan.Key(s.Password),
		tlsDialer: sbtls.NewDialer(N.SystemDialer, tlsConfig),
	}

	transport, err := newTrojanTransportOptions(s)
	if err != nil {
		return nil, err
	}
	if transport.Type != "" {
		rd.transport, err = sbv2ray.NewClientTransport(context.Background(), N.SystemDialer, server, transport, tlsConfig)
		if err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(s.Encryption, "ss;") {
		fields := strings.SplitN(s.Encryption, ";", 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("%w: encryption: %v", dialer.InvalidParameterErr, s.Encryption)
		}
		rd.method, err = ss.CreateMethod(context.Background(), fields[1], ss.MethodOptions{
			Password: fields[2],
		})
		if err != nil {
			return nil, err
		}
	}

	return dialer.NewDialer(rd, true, s.Name, s.Protocol, s.ExportToURL()), nil
}

func (d *sagerNetTrojanDialer) Dial(network, addr string) (net.Conn, error) {
	conn, err := d.dialServer()
	if err != nil {
		return nil, err
	}
	if d.method != nil {
		conn = d.method.DialEarlyConn(conn, d.server)
	}
	destination := M.ParseSocksaddr(addr)
	switch network {
	case "tcp":
		return sbtrojan.NewClientConn(conn, d.key, destination), nil
	case "udp":
		return bufio.NewBindPacketConn(sbtrojan.NewClientPacketConn(conn, d.key), destination), nil
	default:
		_ = conn.Close()
		return nil, net.UnknownNetworkError(network)
	}
}

func (d *sagerNetTrojanDialer) dialServer() (net.Conn, error) {
	if d.transport != nil {
		return d.transport.DialContext(context.Background())
	}
	return d.tlsDialer.DialTLSContext(context.Background(), d.server)
}

func newTrojanTransportOptions(s *Trojan) (option.V2RayTransportOptions, error) {
	var transport option.V2RayTransportOptions
	switch strings.ToLower(s.Type) {
	case "", "none", "origin", "tcp":
	case "ws":
		transport.Type = C.V2RayTransportTypeWebsocket
		transport.WebsocketOptions.Path = s.Path
		if s.Host != "" {
			transport.WebsocketOptions.Headers = badoption.HTTPHeader{
				"Host": badoption.Listable[string]{s.Host},
			}
		}
	case "grpc":
		transport.Type = C.V2RayTransportTypeGRPC
		if s.ServiceName != "" {
			transport.GRPCOptions.ServiceName = s.ServiceName
		} else {
			transport.GRPCOptions.ServiceName = "GunService"
		}
	default:
		return transport, fmt.Errorf("%w: type: %v", dialer.UnexpectedFieldErr, s.Type)
	}
	return transport, nil
}

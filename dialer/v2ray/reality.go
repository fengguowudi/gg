package v2ray

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/mzz2017/gg/dialer"
	ggs "github.com/mzz2017/gg/dialer/sagernet"
	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	sbv2ray "github.com/sagernet/sing-box/transport/v2ray"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common/json/badoption"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	aTLS "github.com/sagernet/sing/common/tls"
)

type realityVLESSDialer struct {
	server    M.Socksaddr
	tls       aTLS.Config
	client    *vless.Client
	transport adapter.V2RayClientTransport
}

func newRealityVLESSDialer(s *V2Ray) (*dialer.Dialer, error) {
	if s.PublicKey == "" {
		return nil, fmt.Errorf("%w: missing reality public key", dialer.InvalidParameterErr)
	}
	serverName := s.SNI
	if serverName == "" {
		serverName = s.Host
	}
	if serverName == "" {
		serverName = s.Add
	}
	tlsConfig, err := ggs.NewRealityTLSConfig(serverName, s.PublicKey, s.ShortID, s.Fingerprint, ggs.SplitCSV(s.Alpn))
	if err != nil {
		return nil, err
	}
	client, err := vless.NewClient(s.ID, s.Flow, ggs.Logger)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(s.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	server := M.ParseSocksaddrHostPort(s.Add, uint16(port))
	rd := &realityVLESSDialer{
		server: server,
		tls:    tlsConfig,
		client: client,
	}
	if s.Net != "" && s.Net != "tcp" {
		rd.transport, err = newRealityTransport(s, server, tlsConfig)
		if err != nil {
			return nil, err
		}
	}
	return dialer.NewDialer(rd, true, s.Ps, s.Protocol, s.ExportToURL()), nil
}

func newRealityTransport(s *V2Ray, server M.Socksaddr, tlsConfig aTLS.Config) (adapter.V2RayClientTransport, error) {
	var transport option.V2RayTransportOptions
	switch s.Net {
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
		transport.GRPCOptions.ServiceName = s.Path
	default:
		return nil, fmt.Errorf("%w: reality network: %v", dialer.UnexpectedFieldErr, s.Net)
	}
	return sbv2ray.NewClientTransport(context.Background(), N.SystemDialer, server, transport, tlsConfig)
}

func (d *realityVLESSDialer) Dial(network, addr string) (net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)
	if d.transport != nil {
		conn, err = d.transport.DialContext(context.Background())
	} else {
		conn, err = N.SystemDialer.DialContext(context.Background(), N.NetworkTCP, d.server)
		if err == nil {
			conn, err = ggs.ClientHandshake(context.Background(), conn, d.tls)
		}
	}
	if err != nil {
		return nil, err
	}
	dst := M.ParseSocksaddr(addr)
	switch network {
	case "tcp":
		return d.client.DialEarlyConn(conn, dst)
	case "udp":
		return d.client.DialEarlyXUDPPacketConn(conn, dst)
	default:
		_ = conn.Close()
		return nil, net.UnknownNetworkError(network)
	}
}

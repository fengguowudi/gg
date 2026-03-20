package httpproxy

import (
	"crypto/tls"
	"net"
	"net/url"
	"strconv"

	"golang.org/x/net/proxy"
)

// Proxy is an HTTP/HTTPS proxy dialer.
type Proxy struct {
	TLSConfig *tls.Config
	Host      string
	HaveAuth  bool
	Username  string
	Password  string
	dialer    proxy.Dialer
}

func New(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	p := &Proxy{
		Host:   u.Host,
		dialer: forward,
	}
	if u.User != nil {
		p.HaveAuth = true
		p.Username = u.User.Username()
		p.Password, _ = u.User.Password()
	}
	if u.Scheme == "https" {
		serverName := u.Query().Get("sni")
		if serverName == "" {
			serverName = u.Hostname()
		}
		skipVerify, _ := strconv.ParseBool(u.Query().Get("allowInsecure"))
		p.TLSConfig = &tls.Config{
			NextProtos:         []string{"h2", "http/1.1"},
			ServerName:         serverName,
			InsecureSkipVerify: skipVerify,
		}
	}
	return p, nil
}

func (p *Proxy) Dial(network, addr string) (net.Conn, error) {
	c, err := p.dialer.Dial("tcp", p.Host)
	if err != nil {
		return nil, err
	}
	if p.TLSConfig != nil {
		c = tls.Client(c, p.TLSConfig)
	}
	return NewConn(c, p, addr), nil
}

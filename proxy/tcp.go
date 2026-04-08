package proxy

import (
	"errors"
	"fmt"
	io2 "github.com/mzz2017/softwind/pkg/zeroalloc/io"
	"io"
	"net"
	"net/netip"
	"time"
)

func (p *Proxy) handleTCP(conn net.Conn) error {
	defer conn.Close()
	loopback, _ := netip.AddrFromSlice(conn.LocalAddr().(*net.TCPAddr).IP)
	tgt := p.GetProjection(loopback)
	if tgt == "" {
		return fmt.Errorf("mapped target address not found: %v", loopback)
	}
	p.log.Tracef("received tcp: %v, tgt: %v", conn.RemoteAddr().String(), tgt)
	c, err := p.dialer.Dial("tcp", tgt)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = RelayTCP(conn, c); err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil // ignore i/o timeout
		}
		return fmt.Errorf("handleTCP relay error: %w", err)
	}
	return nil
}

type WriteCloser interface {
	CloseWrite() error
}

func relayOneWay(dst, src net.Conn, eCh chan<- error) {
	_, err := io2.Copy(dst, src)
	if err == nil || errors.Is(err, io.EOF) {
		if dst, ok := dst.(WriteCloser); ok {
			_ = dst.CloseWrite()
		}
		eCh <- nil
		return
	}

	now := time.Now()
	_ = dst.SetDeadline(now)
	_ = src.SetDeadline(now)
	eCh <- err
}

func RelayTCP(lConn, rConn net.Conn) (err error) {
	eCh := make(chan error, 1)
	go relayOneWay(rConn, lConn, eCh)

	_, e := io2.Copy(lConn, rConn)
	if e == nil || errors.Is(e, io.EOF) {
		if lConn, ok := lConn.(WriteCloser); ok {
			_ = lConn.CloseWrite()
		}
		return <-eCh
	}

	now := time.Now()
	_ = lConn.SetDeadline(now)
	_ = rConn.SetDeadline(now)
	if e != nil {
		<-eCh
		return e
	}
	return <-eCh
}

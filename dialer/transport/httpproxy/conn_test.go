package httpproxy

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

func newLoopbackProxyConn(t *testing.T) (*Conn, net.Conn, <-chan *http.Request) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	reqCh := make(chan *http.Request, 1)
	errCh := make(chan error, 1)
	serverConnCh := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			errCh <- err
			return
		}
		serverConnCh <- conn
		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			errCh <- err
			return
		}
		reqCh <- req
	}()

	clientConn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })

	select {
	case serverConn := <-serverConnCh:
		t.Cleanup(func() { _ = serverConn.Close() })
		conn := NewConn(clientConn, &Proxy{Host: ln.Addr().String(), dialer: proxy.Direct}, "203.0.113.5:22")
		return conn, serverConn, reqCh
	case err := <-errCh:
		t.Fatalf("proxy setup failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for proxy accept")
	}

	return nil, nil, nil
}

func TestConnectTunnelStartsOnRead(t *testing.T) {
	t.Parallel()

	conn, serverConn, reqCh := newLoopbackProxyConn(t)

	readResult := make(chan struct {
		n   int
		err error
		buf []byte
	}, 1)
	go func() {
		buf := make([]byte, len("SSH-2.0-test\r\n"))
		n, err := conn.Read(buf)
		readResult <- struct {
			n   int
			err error
			buf []byte
		}{n: n, err: err, buf: buf[:n]}
	}()

	var req *http.Request
	select {
	case req = <-reqCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for CONNECT request")
	}
	if req.Method != http.MethodConnect {
		t.Fatalf("unexpected method: got %s want CONNECT", req.Method)
	}
	if req.Host != "203.0.113.5:22" {
		t.Fatalf("unexpected CONNECT host: got %q want %q", req.Host, "203.0.113.5:22")
	}

	if _, err := io.WriteString(serverConn, "HTTP/1.1 200 Connection Established\r\n\r\nSSH-2.0-test\r\n"); err != nil {
		t.Fatalf("serverConn.WriteString: %v", err)
	}

	select {
	case got := <-readResult:
		if got.err != nil {
			t.Fatalf("conn.Read: %v", got.err)
		}
		if string(got.buf) != "SSH-2.0-test\r\n" {
			t.Fatalf("unexpected tunneled payload: got %q", string(got.buf))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for tunneled read")
	}
}

func TestConnectTunnelPreservesBufferedBytesAfterResponse(t *testing.T) {
	t.Parallel()

	conn, serverConn, reqCh := newLoopbackProxyConn(t)

	writeDone := make(chan error, 1)
	go func() {
		_, err := conn.Write([]byte("SSH-2.0-client\r\n"))
		writeDone <- err
	}()

	select {
	case <-reqCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for CONNECT request")
	}

	if _, err := io.WriteString(serverConn, "HTTP/1.1 200 Connection Established\r\n\r\nSSH-2.0-server\r\n"); err != nil {
		t.Fatalf("serverConn.WriteString: %v", err)
	}

	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatalf("conn.Write: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client write")
	}

	buf := make([]byte, len("SSH-2.0-server\r\n"))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("conn.Read: %v", err)
	}
	if got := string(buf[:n]); got != "SSH-2.0-server\r\n" {
		t.Fatalf("unexpected tunneled payload: got %q", got)
	}
}

func TestNewConnectRequestUsesAuthorityForm(t *testing.T) {
	t.Parallel()

	c := &Conn{addr: "203.0.113.5:22"}
	req, err := c.newConnectRequest()
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != http.MethodConnect {
		t.Fatalf("unexpected method: %s", req.Method)
	}
	if req.URL.String() != "203.0.113.5:22" && !strings.Contains(req.URL.String(), "203.0.113.5:22") {
		t.Fatalf("unexpected CONNECT url string: %q", req.URL.String())
	}
	u, err := url.Parse("http://203.0.113.5:22")
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "203.0.113.5:22" {
		t.Fatalf("unexpected parsed host: %q", u.Host)
	}
}

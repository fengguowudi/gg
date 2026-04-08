package httpproxy

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

const connectOnReadDelay = 100 * time.Millisecond

type Conn struct {
	net.Conn

	proxy *Proxy
	addr  string

	chShakeFinished chan struct{}
	muShake         sync.Mutex
	reqBuf          io.ReadWriter
	readBuf         *bufio.Reader
	writeSeen       bool
}

func NewConn(c net.Conn, proxy *Proxy, addr string) *Conn {
	return &Conn{
		Conn:            c,
		proxy:           proxy,
		addr:            addr,
		chShakeFinished: make(chan struct{}),
	}
}

func (c *Conn) Write(b []byte) (n int, err error) {
	c.muShake.Lock()
	c.writeSeen = true
	select {
	case <-c.chShakeFinished:
		c.muShake.Unlock()
		return c.Conn.Write(b)
	default:
		defer c.muShake.Unlock()

		_, firstLine, _ := bufio.ScanLines(b, true)
		isHTTPReq := regexp.MustCompile(`^\S+ \S+ HTTP/[\d.]+$`).Match(firstLine)

		if isHTTPReq {
			req, err := c.bufferHTTPRequest(b)
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					return len(b), nil
				}
				return 0, err
			}
			if err = c.writeProxyRequest(req); err != nil {
				return 0, err
			}
			close(c.chShakeFinished)
			return len(b), nil
		}
		if err = c.connectTunnel(); err != nil {
			return 0, err
		}
		return c.Conn.Write(b)
	}
}

func (c *Conn) Read(b []byte) (n int, err error) {
	select {
	case <-c.chShakeFinished:
	default:
		if !c.hasSeenWrite() {
			timer := time.NewTimer(connectOnReadDelay)
			select {
			case <-c.chShakeFinished:
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
			}
		}
		c.muShake.Lock()
		select {
		case <-c.chShakeFinished:
		default:
			if !c.writeSeen {
				if err := c.connectTunnel(); err != nil {
					c.muShake.Unlock()
					return 0, err
				}
			}
		}
		c.muShake.Unlock()
		<-c.chShakeFinished
	}
	if c.readBuf != nil {
		return c.readBuf.Read(b)
	}
	return c.Conn.Read(b)
}

func (c *Conn) hasSeenWrite() bool {
	c.muShake.Lock()
	defer c.muShake.Unlock()
	return c.writeSeen
}

func (c *Conn) bufferHTTPRequest(b []byte) (*http.Request, error) {
	if c.reqBuf == nil {
		c.reqBuf = bytes.NewBuffer(b)
	} else {
		_, _ = c.reqBuf.Write(b)
	}
	req, err := http.ReadRequest(bufio.NewReader(c.reqBuf))
	if err != nil {
		return nil, err
	}
	c.reqBuf = nil
	req.URL.Scheme = "http"
	req.URL.Host = c.addr
	return req, nil
}

func (c *Conn) connectTunnel() error {
	req, err := c.newConnectRequest()
	if err != nil {
		return err
	}
	if err := c.writeProxyRequest(req); err != nil {
		return err
	}

	reader := bufio.NewReader(c.Conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		close(c.chShakeFinished)
		return fmt.Errorf("connect server using proxy error, StatusCode [%d]", resp.StatusCode)
	}
	c.readBuf = reader
	close(c.chShakeFinished)
	return nil
}

func (c *Conn) newConnectRequest() (*http.Request, error) {
	reqURL, err := url.Parse("http://" + c.addr)
	if err != nil {
		return nil, err
	}
	reqURL.Scheme = ""
	return http.NewRequest(http.MethodConnect, reqURL.String(), nil)
}

func (c *Conn) writeProxyRequest(req *http.Request) error {
	req.Close = false
	if c.proxy.HaveAuth {
		token := base64.StdEncoding.EncodeToString([]byte(c.proxy.Username + ":" + c.proxy.Password))
		req.Header.Set("Proxy-Authorization", "Basic "+token)
	}
	if len(req.Header.Values("Proxy-Connection")) > 0 {
		req.Header.Del("Proxy-Connection")
	}
	return req.WriteProxy(c.Conn)
}

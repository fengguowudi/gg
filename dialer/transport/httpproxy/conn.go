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
)

type Conn struct {
	net.Conn

	proxy *Proxy
	addr  string

	chShakeFinished chan struct{}
	muShake         sync.Mutex
	reqBuf          io.ReadWriter
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
	select {
	case <-c.chShakeFinished:
		c.muShake.Unlock()
		return c.Conn.Write(b)
	default:
		defer c.muShake.Unlock()

		_, firstLine, _ := bufio.ScanLines(b, true)
		isHTTPReq := regexp.MustCompile(`^\S+ \S+ HTTP/[\d.]+$`).Match(firstLine)

		var req *http.Request
		if isHTTPReq {
			if c.reqBuf == nil {
				c.reqBuf = bytes.NewBuffer(b)
			} else {
				_, _ = c.reqBuf.Write(b)
			}
			req, err = http.ReadRequest(bufio.NewReader(c.reqBuf))
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					return len(b), nil
				}
				return 0, err
			}
			c.reqBuf = nil
			req.URL.Scheme = "http"
			req.URL.Host = c.addr
		} else {
			reqURL, err := url.Parse("http://" + c.addr)
			if err != nil {
				return 0, err
			}
			reqURL.Scheme = ""
			req, err = http.NewRequest(http.MethodConnect, reqURL.String(), nil)
			if err != nil {
				return 0, err
			}
		}

		req.Close = false
		if c.proxy.HaveAuth {
			token := base64.StdEncoding.EncodeToString([]byte(c.proxy.Username + ":" + c.proxy.Password))
			req.Header.Set("Proxy-Authorization", "Basic "+token)
		}
		if len(req.Header.Values("Proxy-Connection")) > 0 {
			req.Header.Del("Proxy-Connection")
		}

		if err = req.WriteProxy(c.Conn); err != nil {
			return 0, err
		}

		if isHTTPReq {
			close(c.chShakeFinished)
			return len(b), nil
		}

		resp, err := http.ReadResponse(bufio.NewReader(c.Conn), req)
		if err != nil {
			if resp != nil {
				resp.Body.Close()
			}
			return 0, err
		}
		resp.Body.Close()
		close(c.chShakeFinished)
		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("connect server using proxy error, StatusCode [%d]", resp.StatusCode)
		}
		return c.Conn.Write(b)
	}
}

func (c *Conn) Read(b []byte) (n int, err error) {
	<-c.chShakeFinished
	return c.Conn.Read(b)
}

package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/mzz2017/gg/dialer"
	"io"
	"net"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPProxyBasicAuth(t *testing.T) {
	t.Parallel()

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:123"))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()

		req, err := stdhttp.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			done <- err
			return
		}
		if got := req.Header.Get("Proxy-Authorization"); got != wantAuth {
			done <- fmt.Errorf("unexpected proxy auth header: %q", got)
			return
		}
		resp := &stdhttp.Response{
			StatusCode: stdhttp.StatusOK,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(stdhttp.Header),
			Body:       io.NopCloser(strings.NewReader("ok")),
		}
		if err := resp.Write(conn); err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	d, err := NewHTTP("http://admin:123@"+ln.Addr().String(), &dialer.GlobalOption{})
	if err != nil {
		t.Fatal(err)
	}
	client := stdhttp.Client{
		Transport: &stdhttp.Transport{
			DialContext: (&dialer.ContextDialer{Dialer: d.Dialer}).DialContext,
		},
		Timeout: 5 * time.Second,
	}
	req, err := stdhttp.NewRequestWithContext(context.Background(), stdhttp.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %q", string(body))
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

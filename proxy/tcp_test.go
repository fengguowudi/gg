package proxy

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func tcpPair(t *testing.T) (*net.TCPConn, *net.TCPConn) {
	t.Helper()

	ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("ListenTCP: %v", err)
	}
	defer ln.Close()

	acceptCh := make(chan *net.TCPConn, 1)
	errCh := make(chan error, 1)
	go func() {
		conn, err := ln.AcceptTCP()
		if err != nil {
			errCh <- err
			return
		}
		acceptCh <- conn
	}()

	clientConn, err := net.DialTCP("tcp", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatalf("DialTCP: %v", err)
	}

	select {
	case serverConn := <-acceptCh:
		t.Cleanup(func() {
			_ = clientConn.Close()
			_ = serverConn.Close()
		})
		return clientConn, serverConn
	case err := <-errCh:
		_ = clientConn.Close()
		t.Fatalf("AcceptTCP: %v", err)
	case <-time.After(2 * time.Second):
		_ = clientConn.Close()
		t.Fatal("timed out waiting for AcceptTCP")
	}

	return nil, nil
}

func TestRelayTCPPreservesHalfClosedConnections(t *testing.T) {
	t.Parallel()

	leftConn, leftPeer := tcpPair(t)
	rightConn, rightPeer := tcpPair(t)

	relayErrCh := make(chan error, 1)
	go func() {
		relayErrCh <- RelayTCP(leftConn, rightConn)
	}()

	forwarded := []byte("hello")
	if _, err := leftPeer.Write(forwarded); err != nil {
		t.Fatalf("leftPeer.Write: %v", err)
	}
	if err := leftPeer.CloseWrite(); err != nil {
		t.Fatalf("leftPeer.CloseWrite: %v", err)
	}

	if err := rightPeer.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("rightPeer.SetReadDeadline: %v", err)
	}
	gotForwarded := make([]byte, len(forwarded))
	if _, err := io.ReadFull(rightPeer, gotForwarded); err != nil {
		t.Fatalf("rightPeer.ReadFull(forwarded): %v", err)
	}
	if !bytes.Equal(gotForwarded, forwarded) {
		t.Fatalf("forwarded bytes mismatch: got %q want %q", gotForwarded, forwarded)
	}

	reverse := []byte("world")
	if _, err := rightPeer.Write(reverse); err != nil {
		t.Fatalf("rightPeer.Write: %v", err)
	}
	if err := rightPeer.CloseWrite(); err != nil {
		t.Fatalf("rightPeer.CloseWrite: %v", err)
	}

	if err := leftPeer.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("leftPeer.SetReadDeadline: %v", err)
	}
	gotReverse := make([]byte, len(reverse))
	if _, err := io.ReadFull(leftPeer, gotReverse); err != nil {
		t.Fatalf("leftPeer.ReadFull(reverse): %v", err)
	}
	if !bytes.Equal(gotReverse, reverse) {
		t.Fatalf("reverse bytes mismatch: got %q want %q", gotReverse, reverse)
	}

	select {
	case err := <-relayErrCh:
		if err != nil {
			t.Fatalf("RelayTCP: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RelayTCP to finish")
	}
}

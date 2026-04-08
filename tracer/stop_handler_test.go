//go:build linux

package tracer

import (
	"bytes"
	"encoding/binary"
	"net/netip"
	"syscall"
	"testing"
	"unsafe"

	"github.com/mzz2017/gg/proxy"
	"github.com/sirupsen/logrus"
)

func newTestTracer(t *testing.T, ignorePrivateAddr bool) *Tracer {
	t.Helper()

	return &Tracer{
		ignorePrivateAddr: ignorePrivateAddr,
		supportUDP:        true,
		log:               logrus.New(),
		proxy:             proxy.New(logrus.New(), nil),
		storehouse:        MakeStorehouse(),
	}
}

func rawSockaddrInet6(addr netip.Addr, port uint16) []byte {
	sockAddr := RawSockaddrInet6{
		Family: syscall.AF_INET6,
		Addr:   addr.As16(),
	}
	binary.BigEndian.PutUint16(sockAddr.Port[:], port)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&sockAddr)), int(unsafe.Sizeof(sockAddr)))
	return append([]byte(nil), raw...)
}

func TestRestoreConnectRedirectionRestoresOriginalSockaddrAndLength(t *testing.T) {
	t.Parallel()

	tr := &Tracer{
		storehouse: MakeStorehouse(),
	}
	pid := 123
	sockAddrPtr := uintptr(0x1234)
	originalSockAddr := []byte{1, 2, 3, 4, 5}
	originalLen := uint64(32)
	tr.saveConnectRedirection(pid, sockAddrPtr, originalSockAddr, originalLen, 2)

	restorePtracePokeDataFunc := ptracePokeDataFunc
	restorePtraceSetRegsFunc := ptraceSetRegsFunc
	t.Cleanup(func() {
		ptracePokeDataFunc = restorePtracePokeDataFunc
		ptraceSetRegsFunc = restorePtraceSetRegsFunc
	})

	var (
		gotPokePID  int
		gotPokeAddr uintptr
		gotPokeData []byte
		gotLenArg   uint64
	)
	ptracePokeDataFunc = func(pid int, addr uintptr, data []byte) (int, error) {
		gotPokePID = pid
		gotPokeAddr = addr
		gotPokeData = append([]byte(nil), data...)
		return len(data), nil
	}
	ptraceSetRegsFunc = func(pid int, regs *syscall.PtraceRegs) error {
		gotLenArg = Argument(regs, 2)
		return nil
	}

	var regs syscall.PtraceRegs
	setArgument(&regs, 2, 16)
	if err := tr.restoreConnectRedirection(pid, &regs); err != nil {
		t.Fatalf("restoreConnectRedirection: %v", err)
	}
	if gotPokePID != pid {
		t.Fatalf("unexpected pid: got %d want %d", gotPokePID, pid)
	}
	if gotPokeAddr != sockAddrPtr {
		t.Fatalf("unexpected addr: got %#x want %#x", gotPokeAddr, sockAddrPtr)
	}
	if !bytes.Equal(gotPokeData, originalSockAddr) {
		t.Fatalf("unexpected restored sockaddr: got %v want %v", gotPokeData, originalSockAddr)
	}
	if gotLenArg != originalLen {
		t.Fatalf("unexpected restored length: got %d want %d", gotLenArg, originalLen)
	}
	if _, ok := tr.storehouse.Get(pid, syscall.SYS_CONNECT); ok {
		t.Fatal("connect redirection state should be removed after restore")
	}
}

func TestHandleINet6SkipsIPv6Loopback(t *testing.T) {
	t.Parallel()

	tr := newTestTracer(t, false)
	socketInfo := &SocketMetadata{
		Family:   syscall.AF_INET6,
		Type:     syscall.SOCK_STREAM,
		Protocol: syscall.IPPROTO_TCP,
	}
	bSockAddr := rawSockaddrInet6(netip.MustParseAddr("::1"), 443)

	got, err := tr.handleINet6(socketInfo, bSockAddr)
	if err != nil {
		t.Fatalf("handleINet6: %v", err)
	}
	if got != nil {
		t.Fatalf("expected loopback address to be skipped, got %v", got)
	}
}

func TestHandleINet6SkipsIPv4MappedPrivateAddress(t *testing.T) {
	t.Parallel()

	tr := newTestTracer(t, true)
	socketInfo := &SocketMetadata{
		Family:   syscall.AF_INET6,
		Type:     syscall.SOCK_STREAM,
		Protocol: syscall.IPPROTO_TCP,
	}
	bSockAddr := rawSockaddrInet6(netip.MustParseAddr("::ffff:192.168.1.20"), 443)

	got, err := tr.handleINet6(socketInfo, bSockAddr)
	if err != nil {
		t.Fatalf("handleINet6: %v", err)
	}
	if got != nil {
		t.Fatalf("expected IPv4-mapped private address to be skipped, got %v", got)
	}
}

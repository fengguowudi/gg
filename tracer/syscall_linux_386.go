//go:build linux && 386

package tracer

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

const (
	socketCallSocket  = 1
	socketCallConnect = 3
	socketCallSendto  = 11
	socketCallSendmsg = 16
)

// RawSockaddrInet4 is a bit different from syscall.RawSockaddrInet4 that Port should be encoded by BigEndian.
type RawSockaddrInet4 struct {
	Family uint16
	Port   [2]byte
	Addr   [4]byte /* in_addr */
	Zero   [8]uint8
}

// RawSockaddrInet6 is a bit different from syscall.RawSockaddrInet6 that fields except Family should be encoded by BigEndian.
type RawSockaddrInet6 struct {
	Family   uint16
	Port     [2]byte
	Flowinfo [4]byte
	Addr     [16]byte /* in6_addr */
	Scope_id [4]byte
}

type RawMsgHdr struct {
	MsgName       uint32
	LenMsgName    uint32
	MsgIov        uint32
	LenMsgIov     uint32
	MsgControl    uint32
	LenMsgControl uint32
	Flags         int32
}

func arguments(regs *syscall.PtraceRegs) []uint64 {
	return []uint64{
		uint64(uint32(regs.Ebx)),
		uint64(uint32(regs.Ecx)),
		uint64(uint32(regs.Edx)),
		uint64(uint32(regs.Esi)),
		uint64(uint32(regs.Edi)),
		uint64(uint32(regs.Ebp)),
	}
}

func setArgument(regs *syscall.PtraceRegs, order int, val uint64) {
	switch order {
	case 0:
		regs.Ebx = int32(val)
	case 1:
		regs.Ecx = int32(val)
	case 2:
		regs.Edx = int32(val)
	case 3:
		regs.Esi = int32(val)
	case 4:
		regs.Edi = int32(val)
	case 5:
		regs.Ebp = int32(val)
	}
}

func returnValueInt(regs *syscall.PtraceRegs) (int, syscall.Errno) {
	if regs.Eax < 0 {
		return int(regs.Eax), syscall.Errno(-regs.Eax)
	}
	return int(regs.Eax), 0
}

func isEntryStop(regs *syscall.PtraceRegs) bool {
	return regs.Eax == -int32(syscall.ENOSYS)
}

func inst(regs *syscall.PtraceRegs) int {
	return int(regs.Orig_eax)
}

func ptraceSetRegs(pid int, regs *syscall.PtraceRegs) error {
	return syscall.PtraceSetRegs(pid, regs)
}

func ptraceGetRegs(pid int, regs *syscall.PtraceRegs) error {
	return syscall.PtraceGetRegs(pid, regs)
}

func decodeSyscall(pid int, regs *syscall.PtraceRegs) (int, []uint64, error) {
	actual := inst(regs)
	switch actual {
	case syscall.SYS_SOCKETCALL:
		call := uint32(regs.Ebx)
		argsPtr := uintptr(uint32(regs.Ecx))
		switch call {
		case socketCallSocket:
			args, err := readSocketcallArgs(pid, argsPtr, 3)
			return traceSysSocket, args, err
		case socketCallConnect:
			args, err := readSocketcallArgs(pid, argsPtr, 3)
			return traceSysConnect, args, err
		case socketCallSendto:
			args, err := readSocketcallArgs(pid, argsPtr, 6)
			return traceSysSendto, args, err
		case socketCallSendmsg:
			args, err := readSocketcallArgs(pid, argsPtr, 3)
			return traceSysSendmsg, args, err
		default:
			return traceSysUnknown, arguments(regs), nil
		}
	case syscall.SYS_FCNTL, syscall.SYS_FCNTL64:
		return traceSysFcntl, arguments(regs), nil
	case syscall.SYS_CLOSE:
		return traceSysClose, arguments(regs), nil
	default:
		return traceSysUnknown, arguments(regs), nil
	}
}

func readArgumentValue(pid int, regs *syscall.PtraceRegs, order int) (uint64, error) {
	if inst(regs) != syscall.SYS_SOCKETCALL {
		return Argument(regs, order), nil
	}
	return readSocketcallArgument(pid, uintptr(uint32(regs.Ecx)), order)
}

func writeArgumentValue(pid int, regs *syscall.PtraceRegs, order int, val uint64) error {
	if inst(regs) != syscall.SYS_SOCKETCALL {
		newRegs := *regs
		setArgument(&newRegs, order, val)
		return ptraceSetRegsFunc(pid, &newRegs)
	}
	return writeSocketcallArgument(pid, uintptr(uint32(regs.Ecx)), order, val)
}

func readSocketcallArgs(pid int, argsPtr uintptr, count int) ([]uint64, error) {
	args := make([]uint64, count)
	for i := 0; i < count; i++ {
		v, err := readSocketcallArgument(pid, argsPtr, i)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return args, nil
}

func readSocketcallArgument(pid int, argsPtr uintptr, order int) (uint64, error) {
	buf := make([]byte, 4)
	if _, err := syscall.PtracePeekData(pid, argsPtr+uintptr(order*4), buf); err != nil {
		return 0, fmt.Errorf("read socketcall argument %d: %w", order, err)
	}
	return uint64(binary.LittleEndian.Uint32(buf)), nil
}

func writeSocketcallArgument(pid int, argsPtr uintptr, order int, val uint64) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(val))
	if _, err := ptracePokeDataFunc(pid, argsPtr+uintptr(order*4), buf); err != nil {
		return fmt.Errorf("write socketcall argument %d: %w", order, err)
	}
	return nil
}

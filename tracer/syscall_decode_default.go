//go:build linux && !386

package tracer

import "syscall"

func decodeSyscall(pid int, regs *syscall.PtraceRegs) (int, []uint64, error) {
	actual := inst(regs)
	args := arguments(regs)
	switch actual {
	case syscall.SYS_SOCKET:
		return traceSysSocket, args, nil
	case syscall.SYS_CONNECT:
		return traceSysConnect, args, nil
	case syscall.SYS_SENDTO:
		return traceSysSendto, args, nil
	case syscall.SYS_SENDMSG:
		return traceSysSendmsg, args, nil
	case syscall.SYS_FCNTL:
		return traceSysFcntl, args, nil
	case syscall.SYS_CLOSE:
		return traceSysClose, args, nil
	default:
		return traceSysUnknown, args, nil
	}
}

func readArgumentValue(pid int, regs *syscall.PtraceRegs, order int) (uint64, error) {
	return Argument(regs, order), nil
}

func writeArgumentValue(pid int, regs *syscall.PtraceRegs, order int, val uint64) error {
	newRegs := *regs
	setArgument(&newRegs, order, val)
	return ptraceSetRegsFunc(pid, &newRegs)
}

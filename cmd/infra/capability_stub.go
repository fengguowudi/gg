//go:build !linux

package infra

import "fmt"

var (
	ErrGetPtraceScope     = fmt.Errorf("error when get ptrace scope")
	ErrGetCapability      = fmt.Errorf("error when get capability")
	ErrBadPtraceScope     = fmt.Errorf("bad ptrace scope")
	ErrBadCapability      = fmt.Errorf("bad capability")
	ErrUnknownPtraceScope = fmt.Errorf("unknown ptrace scope")
)

func GetPtraceScope() (int, error) {
	return 0, nil
}

func CheckPtraceCapability() error {
	return nil
}

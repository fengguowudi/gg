//go:build !linux

package tracer

import (
	"context"
	"fmt"
	"os"

	"github.com/mzz2017/gg/dialer"
	"github.com/sirupsen/logrus"
)

var ErrUnsupportedPlatform = fmt.Errorf("tracer is only supported on linux")

type Tracer struct{}

func New(ctx context.Context, name string, argv []string, attr *os.ProcAttr, dialer *dialer.Dialer, ignoreUDP bool, ignorePrivateAddr bool, logger *logrus.Logger) (*Tracer, error) {
	return nil, ErrUnsupportedPlatform
}

func (t *Tracer) Wait() (exitCode int, err error) {
	return 1, ErrUnsupportedPlatform
}

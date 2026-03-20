package sagernet

import (
	"context"
	"strings"

	"github.com/mzz2017/gg/common"
	singLogger "github.com/sagernet/sing/common/logger"
)

type NopLogger struct{}

func (NopLogger) Trace(args ...any)                             {}
func (NopLogger) Debug(args ...any)                             {}
func (NopLogger) Info(args ...any)                              {}
func (NopLogger) Warn(args ...any)                              {}
func (NopLogger) Error(args ...any)                             {}
func (NopLogger) Fatal(args ...any)                             {}
func (NopLogger) Panic(args ...any)                             {}
func (NopLogger) TraceContext(ctx context.Context, args ...any) {}
func (NopLogger) DebugContext(ctx context.Context, args ...any) {}
func (NopLogger) InfoContext(ctx context.Context, args ...any)  {}
func (NopLogger) WarnContext(ctx context.Context, args ...any)  {}
func (NopLogger) ErrorContext(ctx context.Context, args ...any) {}
func (NopLogger) FatalContext(ctx context.Context, args ...any) {}
func (NopLogger) PanicContext(ctx context.Context, args ...any) {}

var Logger singLogger.ContextLogger = NopLogger{}

func SplitCSV(value string) []string {
	if value == "" {
		return nil
	}
	fields := strings.Split(value, ",")
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			result = append(result, field)
		}
	}
	return result
}

func QueryBool(values map[string][]string, keys ...string) bool {
	for _, key := range keys {
		if len(values[key]) == 0 {
			continue
		}
		if common.StringToBool(values[key][0]) {
			return true
		}
	}
	return false
}

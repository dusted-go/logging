package stackdriver

import (
	"fmt"
	"runtime"
	"strings"
)

type Stack []uintptr

func (s *Stack) String() string {
	sb := strings.Builder{}
	frames := runtime.CallersFrames(*s)
	for {
		f, more := frames.Next()
		if strings.HasSuffix(f.File, "stackdriver/stacktrace.go") {
			continue
		}
		sb.WriteString(
			fmt.Sprintf("\nat %s:%d\n   --> %s", f.File, f.Line, f.Function),
		)
		if !more {
			return sb.String()
		}
	}
}

func (s *Stack) Slice() []string {
	var out []string
	frames := runtime.CallersFrames(*s)
	for {
		f, more := frames.Next()
		if strings.HasPrefix(f.Function, "log/slog.") {
			continue
		}

		out = append(out, fmt.Sprintf("%s:%d (%s)", f.File, f.Line, f.Function))
		if !more {
			return out
		}
	}
}

func CaptureStack() *Stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var s Stack = pcs[0:n]
	return &s
}

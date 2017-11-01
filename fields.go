package jolt

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Fields map[string]interface{}

type FieldFunc func() string

// Location logs the source file and line number of the Logger.Print
// caller.
//
// If the error is wrapped by github.com/pkg/errors then then the source
// file and line number will be based upon when the original error was
// wrapped.
func Location() FieldFunc {
	return func() string {
		return location()
	}
}

func Package() FieldFunc {
	return func() string {
		return pkgName()
	}
}

var stacktracePool = sync.Pool{
	New: func() interface{} {
		return make([]uintptr, 64)
	},
}

func Stacktrace() []runtime.Frame {
	pcs := stacktracePool.Get().([]uintptr)
	var numFrames int
	for {
		numFrames = runtime.Callers(2, pcs)
		if numFrames < len(pcs) {
			break
		}
		pcs = make([]uintptr, len(pcs)*2)
	}
	trace := make([]runtime.Frame, 0)
	frames := runtime.CallersFrames(pcs[:numFrames])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if strings.Contains(frame.Function, "jolt.") {
			continue
		}
		trace = append(trace, frame)
	}
	return trace
}

func callerFrame() *runtime.Frame {
	frames := Stacktrace()
	if len(frames) == 0 {
		return nil
	}
	return &frames[0]
}

func location() string {
	frame := callerFrame()
	if frame == nil {
		return "unknown"
	}
	fp := filepath.Base(filepath.Dir(frame.File))
	filename := filepath.Base(frame.File)
	return fmt.Sprintf("%s:%d", filepath.Join(fp, filename), frame.Line)
}

func funcName() (s string) {
	frame := callerFrame()
	if frame == nil {
		return
	}
	parts := strings.Split(frame.Function, ".")
	return parts[len(parts)-1]
}

func fullPkgName() (s string) {
	frame := callerFrame()
	if frame == nil {
		return
	}
	parts := strings.Split(frame.Function, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func pkgName() string {
	return filepath.Base(fullPkgName())
}

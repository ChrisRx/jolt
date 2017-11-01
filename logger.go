package jolt

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// A Logger provides simple, user-friendly and thread-safe logging.
type Logger struct {
	defaults Fields

	mu sync.Mutex
	w  io.Writer
}

// New initializes a new *jolt.Logger. By default, If no writers
// are provided os.Stderr is used.
func New(w ...io.Writer) *Logger {
	if len(w) == 0 {
		w = []io.Writer{os.Stderr}
	}
	if len(w) > 1 {
		w[0] = io.MultiWriter(w...)
	}
	return &Logger{
		w:        w[0],
		defaults: make(Fields),
	}
}

// DefaultLogger constructs a new *jolt.Logger with the following
// defaults:
//
//     ts     current UTC time (RFC3339)
//     loc    the source file and line number of the caller
func DefaultLogger(w ...io.Writer) *Logger {
	l := New(w...)
	l.With(Fields{
		"ts": func() string {
			return time.Now().UTC().Format(time.RFC3339)
		},
		"loc": Location(),
	})
	return l
}

// With allows default Fields to be set. If a field is specified it will
// be in every invocation of Logger.Print.
func (j *Logger) With(m Fields) *Logger {
	l := j.clone()
	for k, v := range m {
		l.defaults[k] = v
	}
	return l
}

func (j *Logger) clone() *Logger {
	l := &Logger{
		defaults: j.defaults,
		w:        j.w,
	}
	return l
}

type stacktracer interface {
	StackTrace() errors.StackTrace
}

func (j *Logger) Print(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	for i, a := range args {
		switch t := a.(type) {
		case Fields:
			j.printFields(t)
			args = append(args[:i], args[i+1:]...)
		case map[string]interface{}:
			j.printFields(t)
			args = append(args[:i], args[i+1:]...)
		}
	}
	if len(args) > 0 {
		var format string
		switch t := args[0].(type) {
		case string:
			format = t
		case fmt.Stringer:
			format = t.String()
		case error:
			if _, ok := t.(stacktracer); !ok {
				args[0] = errors.WithStack(t)
			}
			return j.print(args...)
		default:
			panic(fmt.Errorf("received invalid type '%v' in arguments", reflect.TypeOf(args[0])))
		}
		return j.print(fmt.Sprintf(format, args[1:]...))
	}
	return nil
}

func (j *Logger) print(a ...interface{}) error {
	fields := make(Fields)
	for i := 0; i < len(a); i++ {
		if err, ok := a[i].(stacktracer); ok {
			frames := err.StackTrace()
			if len(frames) == 0 {
				continue
			}
			frame := frames[0]
			s := fmt.Sprintf("%+v", frame)
			parts := strings.Split(s, "\t")
			if len(parts) < 2 {
				continue
			}
			fp := filepath.Base(filepath.Dir(parts[1]))
			filename := filepath.Base(parts[1])
			fields["loc"] = filepath.Join(fp, filename)
		}
	}
	fields["msg"] = fmt.Sprint(a...)
	return j.printFields(fields)
}

func (j *Logger) printFields(m Fields) error {
	tmp := make(Fields)
	for k, v := range j.defaults {
		switch t := v.(type) {
		case func() string:
			tmp[k] = t()
		case FieldFunc:
			tmp[k] = t()
		default:
			tmp[k] = v
		}
	}
	for k, v := range m {
		tmp[k] = v
	}
	b, err := json.Marshal(tmp)
	if err != nil {
		panic(err)
	}
	b = append(b, '\n')
	j.mu.Lock()
	defer j.mu.Unlock()
	_, err = j.w.Write(b)
	if err != nil {
		return err
	}
	return nil
}

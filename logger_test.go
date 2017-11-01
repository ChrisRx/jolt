package jolt_test

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/ChrisRx/jolt"
)

func TestInvalidPrintArgsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("test did not panic")
		}
	}()
	j := jolt.New(os.Stdout)
	j.Print(0)
}

func newLogger(w ...io.Writer) *jolt.Logger {
	j := jolt.New(w...)
	j.With(jolt.Fields{
		"ts": func() string {
			var t time.Time
			return t.Format(time.RFC3339)
		},
	})
	return j
}

func lineno(tag string) (int, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return 0, errors.Errorf("cannot determine current filepath")
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", data, parser.ParseComments)
	if err != nil {
		return 0, err
	}
	for _, group := range f.Comments {
		for _, comment := range group.List {
			if strings.TrimPrefix(comment.Text, "// ") == tag {
				return fset.Position(comment.Pos()).Line, nil
			}
		}
	}
	return 0, errors.Errorf("unable to find tag: %#v", tag)
}

func TestStackTraceSkipFrames(t *testing.T) {
	var buf bytes.Buffer
	j := jolt.New(&buf).With(jolt.Fields{
		"loc": jolt.Location(),
		"pkg": jolt.Package(),
	})
	err := j.Print("test") // n:0
	if err != nil {
		t.Error(err)
	}
	n, err := lineno("n:0")
	if err != nil {
		t.Fatal(err)
	}
	Expected := fmt.Sprintf("{\"loc\":\"jolt/logger_test.go:%d\",\"msg\":\"test\",\"pkg\":\"jolt_test\"}\n", n)
	if diff := cmp.Diff(buf.String(), Expected); diff != "" {
		t.Errorf("after Print: (-got +want)\n%s", diff)
	}
}

func TestPkgErrorsTrace(t *testing.T) {
	var buf bytes.Buffer
	j := jolt.New(&buf).With(jolt.Fields{
		"loc": jolt.Location(),
		"pkg": jolt.Package(),
	})
	test := errors.Errorf("test") // n:1
	err := j.Print(test)
	if err != nil {
		t.Error(err)
	}
	n, err := lineno("n:1")
	if err != nil {
		t.Fatal(err)
	}
	Expected := fmt.Sprintf("{\"loc\":\"jolt/logger_test.go:%d\",\"msg\":\"test\",\"pkg\":\"jolt_test\"}\n", n)
	if diff := cmp.Diff(buf.String(), Expected); diff != "" {
		t.Errorf("after Print: (-got +want)\n%s", diff)
	}
}

func TestJoltFields(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expected interface{}
	}{
		{
			Input:    jolt.Fields{"hostname": "joltbox", "errorcode": 9001},
			Expected: "{\"errorcode\":9001,\"hostname\":\"joltbox\",\"ts\":\"0001-01-01T00:00:00Z\"}\n",
		},
		{
			Input:    map[string]interface{}{"hostname": "joltbox", "errorcode": 9001},
			Expected: "{\"errorcode\":9001,\"hostname\":\"joltbox\",\"ts\":\"0001-01-01T00:00:00Z\"}\n",
		},
	}

	for i, tc := range cases {
		var buf bytes.Buffer
		j := newLogger(&buf)
		j.Print(tc.Input)
		if diff := cmp.Diff(buf.String(), tc.Expected); diff != "" {
			t.Errorf("%d: after Print: (-got +want)\n%s", i, diff)
		}
	}
}

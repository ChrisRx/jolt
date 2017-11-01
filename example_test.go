package jolt_test

import (
	"os"
	"time"

	"github.com/ChrisRx/jolt"
)

func ExampleJoltPrint() {
	j := jolt.New(os.Stdout)
	j.Print("jolt'n like a sultan")
	//Output: {"msg":"jolt'n like a sultan"}
}

func ExampleJoltPrintf() {
	j := jolt.New(os.Stdout)
	j.Print("%s'n like a sultan", "jolt")
	//Output: {"msg":"jolt'n like a sultan"}
}

func ExampleJoltFields() {
	j := jolt.New(os.Stdout)
	j.Print(jolt.Fields{
		"code": 9001,
		"msg":  "jolt'n like a sultan",
	})
	//Output: {"code":9001,"msg":"jolt'n like a sultan"}
}

func ExampleJoltDefaultFields() {
	j := jolt.New(os.Stdout).With(jolt.Fields{
		"ts": func() string {
			var t time.Time
			return t.Format(time.RFC3339)
		},
	})
	j.Print(jolt.Fields{
		"msg": "jolt'n like a sultan",
	})
	//Output: {"msg":"jolt'n like a sultan","ts":"0001-01-01T00:00:00Z"}
}

package main

import (
	"os"

	"labs.lesiw.io/ops/goapp"
	"lesiw.io/ops"
)

func main() {
	goapp.Name = "moxie"
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "check")
	}
	ops.Handle(goapp.Ops{})
}

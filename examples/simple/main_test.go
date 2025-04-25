package main

// Make sure to run `go generate` to generate the moxie files

import (
	"testing"
)

func Test_Person_with_moxie(t *testing.T) {
	stub := new(StubForWorker)
	stub._IsWorkingWell_Return(false)
	company := Company{}
	company.Workers = []Worker{stub}
	want := false
	got := company.IsGonnaBeRich()
	if got != want {
		t.Errorf("Mocking failed, got %v, want %v", got, want)
	}
}

package main

// Make sure to run `go generate` to generate the moxie files

import (
	"testing"
)

func Test_Person_with_moxie(t *testing.T) {
	p := new(Person)
	want := false
	(*p)._IsAlive_Return(want)

	got := p.IsAlive()
	if got != want {
		t.Errorf("Mocking failed, got %v, want %v", got, want)
	}
}

package testdata

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"lesiw.io/moxie/internal/testdata/pkg"
)

func TestNoMock(t *testing.T) {
	t.Cleanup(func() { pkg.SimpleCalled = false })
	var m0 M0
	m0.Simple()
	if !pkg.SimpleCalled {
		t.Error("want Simple() call")
	}
}

func TestStub(t *testing.T) {
	t.Cleanup(func() { pkg.SimpleCalled = false })
	var m0 M0
	m0._Simple_Stub()
	m0.Simple()
	if pkg.SimpleCalled {
		t.Error("want mock, got Simple() call")
	}
}

func TestOneResult(t *testing.T) {
	var m0 M0
	want := errors.New("error result")
	m0._OneResult_Return(want)
	if got := m0.OneResult(); want != got {
		t.Errorf("M0.OneResult() call #1: want %q, got %q", want, got)
	}
	// Calling a second time should return the same result.
	if got := m0.OneResult(); want != got {
		t.Errorf("M0.OneResult() call #2: want %q, got %q", want, got)
	}
}

func TestOneResultQueue(t *testing.T) {
	var m0 M0
	err1 := errors.New("error one")
	err2 := errors.New("error two")
	m0._OneResult_Return(err1)
	m0._OneResult_Return(err2)
	if got := m0.OneResult(); err1 != got {
		t.Errorf("M0.OneResult() call #1: want %q, got %q", err1, got)
	}
	if got := m0.OneResult(); err2 != got {
		t.Errorf("M0.OneResult() call #2: want %q, got %q", err1, got)
	}
	if got := m0.OneResult(); err2 != got {
		t.Errorf("M0.OneResult() call #3: want %q, got %q", err1, got)
	}
}

func TestCalls(t *testing.T) {
	var m0 M0
	m0.OneParamNoResult("call one")
	m0.OneParamNoResult("call two")
	m0.OneParamNoResult("call three")
	want := []_M0_OneParamNoResult_Call{
		{"call one"},
		{"call two"},
		{"call three"},
	}
	got := m0._OneParamNoResult_Calls()
	opt := cmpopts.EquateComparable(_M0_OneParamNoResult_Call{})
	if !cmp.Equal(want, got, opt) {
		t.Errorf("M0._OneParamNoResult_Calls():\n%s", cmp.Diff(want, got, opt))
	}
}

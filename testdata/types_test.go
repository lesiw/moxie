package testdata

import (
	"errors"
	"testing"
)

func TestNoMock(t *testing.T) {
	m0 := &M0{new(T0)}
	m0.Simple()
	if !m0.simpleCalled {
		t.Error("want Simple() call")
	}
}

func TestStub(t *testing.T) {
	m0 := &M0{new(T0)}
	m0._Simple_Stub()
	m0.Simple()
	if m0.simpleCalled {
		t.Error("want mock, got Simple() call")
	}
}

func TestOneResult(t *testing.T) {
	m0 := &M0{new(T0)}
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
	m0 := &M0{new(T0)}
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

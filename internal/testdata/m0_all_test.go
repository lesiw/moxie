package testdata

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"lesiw.io/moxie/internal/testdata/pkg"
)

func TestStubAll(t *testing.T) {
	var m0 M0
	t.Cleanup(func() { pkg.SimpleCalled = false })
	t.Run("TestStubAllSubTest", func(t *testing.T) {
		t.Cleanup(func() { pkg.SimpleCalled = false })
		new(M0)._Simple_StubAll(t)
		m0.Simple()
		if pkg.SimpleCalled {
			t.Error("want mock, got Simple() call")
		}
	})
	m0.Simple()
	if !pkg.SimpleCalled {
		t.Errorf("want Simple() call, got mock")
	}
}

func TestUnmockAll(t *testing.T) {
	var m0 M0
	t.Cleanup(func() { pkg.SimpleCalled = false })
	t.Run("TestUnmockAllSubTest", func(t *testing.T) {
		t.Cleanup(func() { pkg.SimpleCalled = false })
		new(M0)._Simple_StubAll(t)
		m0.Simple()
		if pkg.SimpleCalled {
			t.Error("want mock, got Simple() call")
		}
		new(M0)._Simple_DoAll(t, nil)
		m0.Simple()
		if !pkg.SimpleCalled {
			t.Error("want Simple() call, got mock")
		}
	})
	m0.Simple()
	if !pkg.SimpleCalled {
		t.Errorf("want Simple() call, got mock")
	}
}

func TestOneResultAll(t *testing.T) {
	var m0 M0
	t.Run("TestOneResultAllSubTest", func(t *testing.T) {
		want := errors.New("error result")
		new(M0)._OneResult_ReturnAll(t, want)
		if got := m0.OneResult(); want != got {
			t.Errorf("M0.OneResult() call #1: want %v, got %v", want, got)
		}
		// Calling a second time should return the same result.
		if got := m0.OneResult(); want != got {
			t.Errorf("M0.OneResult() call #2: want %v, got %v", want, got)
		}
	})
	// Calling outside the test scope should return nil
	// (the actual return of the function).
	if got, want := m0.OneResult(), error(nil); want != got {
		t.Errorf("M0.OneResult() call #3: want %v, got %v", want, got)
	}
}

func TestOneResultAllQueue(t *testing.T) {
	var m0 M0
	t.Run("TestOneResultAllQueueSubTest", func(t *testing.T) {
		err1 := errors.New("error one")
		err2 := errors.New("error two")
		new(M0)._OneResult_ReturnAll(t, err1)
		new(M0)._OneResult_ReturnAll(t, err2)
		if got := m0.OneResult(); err1 != got {
			t.Errorf("M0.OneResult() call #1: want %v, got %v", err1, got)
		}
		if got := m0.OneResult(); err2 != got {
			t.Errorf("M0.OneResult() call #2: want %v, got %v", err1, got)
		}
		if got := m0.OneResult(); err2 != got {
			t.Errorf("M0.OneResult() call #3: want %v, got %v", err1, got)
		}
	})
	if got, want := m0.OneResult(), error(nil); want != got {
		t.Errorf("M0.OneResult() call #3: want %v, got %v", want, got)
	}
}

func TestAllCalls(t *testing.T) {
	new(M0)._OneParamNoResult_ResetAllCalls()
	t.Run("TestAllCallsSubTest", func(t *testing.T) {
		t.Cleanup(new(M0)._OneParamNoResult_ResetAllCalls)
		new(M0).OneParamNoResult("call one")
		new(M0).OneParamNoResult("call two")
		new(M0).OneParamNoResult("call three")
		want := []_M0_OneParamNoResult_Call{
			{"call one"},
			{"call two"},
			{"call three"},
		}
		got := new(M0)._OneParamNoResult_AllCalls()
		opt := cmpopts.EquateComparable(_M0_OneParamNoResult_Call{})
		if !cmp.Equal(want, got, opt) {
			t.Errorf("M0._OneParamNoResult_AllCalls():\n%s",
				cmp.Diff(want, got, opt))
		}
	})
	if got, want := len(new(M0)._OneParamNoResult_AllCalls()), 0; got != want {
		t.Errorf("M0._OneParamNoResult_AllCalls() did not reset")
	}
}

func TestAllInterface(t *testing.T) {
	var m0 M0
	new(M0)._Read_StubAll(t)
	_, _ = m0.Read(nil) // Validate this does not panic.
}

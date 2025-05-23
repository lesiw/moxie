package pkg

import "io"

// The type under test cannot be the empty struct.
// Empty structs are subject to special optimization and may return the same
// pointer receiver between different tests.
type T0 struct {
	_             bool
	io.ReadWriter // Double interface embed.
}

type String string
type Int int

var SimpleCalled bool

func (T0) Simple() { SimpleCalled = true }

func (T0) OneResult() error                    { return nil }
func (T0) OneNamedResult() (err error)         { return }
func (T0) TwoResults() (Int, error)            { return 0, nil }
func (T0) TwoNamedResults() (n Int, err error) { return }

func (T0) OneParamNoResult(String)                {}
func (T0) OneParamOneResult(String) error         { return nil }
func (T0) OneParamTwoResults(String) (Int, error) { return 0, nil }

func (T0) TwoParamsNoResult(String, String)                {}
func (T0) TwoParamsOneResult(String, String) error         { return nil }
func (T0) TwoParamsTwoResults(String, String) (Int, error) { return 0, nil }

func (T0) VariadicNoResult(...String)                {}
func (T0) VariadicOneResult(...String) error         { return nil }
func (T0) VariadicTwoResults(...String) (Int, error) { return 0, nil }

func (T0) MixedNoResult(String, ...String)                {}
func (T0) MixedOneResult(String, ...String) error         { return nil }
func (T0) MixedTwoResults(String, ...String) (Int, error) { return 0, nil }

func (T0) NamedParamNoResult(x String)                {}
func (T0) NamedParamOneResult(x String) error         { return nil }
func (T0) NamedParamTwoResults(x String) (Int, error) { return 0, nil }

func (T0) NamedMixedNoResult(x String, y ...String)                {}
func (T0) NamedMixedOneResult(x String, y ...String) error         { return nil }
func (T0) NamedMixedTwoResults(x String, y ...String) (Int, error) { return 0, nil }

func (T0) AllNamedIdentifiers(x String, y ...String) (n Int, err error) { return }

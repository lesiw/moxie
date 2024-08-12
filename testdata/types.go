package testdata

type M0 struct{ *T0 }

type T0 struct {
	simpleCalled bool
}

func (t *T0) Simple() { t.simpleCalled = true }

func (T0) OneResult() error                    { return nil }
func (T0) OneNamedResult() (err error)         { return }
func (T0) TwoResults() (int, error)            { return 0, nil }
func (T0) TwoNamedResults() (n int, err error) { return }

func (T0) OneParamNoResult(string)                {}
func (T0) OneParamOneResult(string) error         { return nil }
func (T0) OneParamTwoResults(string) (int, error) { return 0, nil }

func (T0) TwoParamsNoResult(string, string)                {}
func (T0) TwoParamsOneResult(string, string) error         { return nil }
func (T0) TwoParamsTwoResults(string, string) (int, error) { return 0, nil }

func (T0) VariadicNoResult(...string)                {}
func (T0) VariadicOneResult(...string) error         { return nil }
func (T0) VariadicTwoResults(...string) (int, error) { return 0, nil }

func (T0) MixedNoResult(string, ...string)                {}
func (T0) MixedOneResult(string, ...string) error         { return nil }
func (T0) MixedTwoResults(string, ...string) (int, error) { return 0, nil }

func (T0) NamedParamNoResult(x string)                {}
func (T0) NamedParamOneResult(x string) error         { return nil }
func (T0) NamedParamTwoResults(x string) (int, error) { return 0, nil }

func (T0) NamedMixedNoResult(x string, y ...string)                {}
func (T0) NamedMixedOneResult(x string, y ...string) error         { return nil }
func (T0) NamedMixedTwoResults(x string, y ...string) (int, error) { return 0, nil }

func (T0) AllNamedIdentifiers(x string, y ...string) (n int, err error) { return }

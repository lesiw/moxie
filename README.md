# lesiw.io/moxie: Generate mocks with proxy functions

Go has a convenient feature called [embedding][embedding]. Method calls are
transparently passed through to embedded types.

`moxie` generates proxy methods that are only added during `go test`.

## Usage

Add this line to your code, where `T` is the type you want to instrument, and
`E` is the embedded type whose functions you want to intercept.

``` go
//go:generate go run lesiw.io/moxie@latest T
type T struct {
    E
}
```

Then run `go generate`.

## Functions

`moxie` makes the following functions available at test time.

### Mock

```
(*T)._Func_Stub()           -> when called, return zero values.
(*T)._Func_Return(...)      -> when called, return these values.
(*T)._Func_Do(func() {...}) -> when called, run this instead.
                               calling this with nil will un-mock the function.
```

Mocks queue. For example, mocking with `(*T)._Func_Return` twice will return the
first set of values the next time the function is called, then return the second
set of values in subsequent calls.

### Inspect

```
(*T)._Func_Calls() []_T_Func_Call -> return calls to Func.
```

[embedding]: https://go.dev/doc/effective_go#embedding

# lesiw.io/moxie: Mock by proxy

Go has a convenient feature called [embedding][embedding]. Method calls are
transparently passed through to embedded types.

`moxie` generates proxy methods that are only added during `go test`.

## Usage

Add this line to your code, where `T` is the type you want to instrument, and
`*E` is the embedded type whose functions you want to intercept.

``` go
//go:generate go run lesiw.io/moxie@latest T
type T struct {
    *E
}
```

Then run `go generate`.

## Functions

`moxie` makes the following functions available at test time.

```
(*T)._Func_Patch()           -> patch Func, returning zero values.
(*T)._Func_Return(r0, ...)   -> patch Func, returning parameters.
(*T)._Func_Mock(mock func()) -> override Func behavior.
(*T)._Func_Calls() []Calls   -> return calls to Func.
```

[embedding]: https://go.dev/doc/effective_go#embedding

# lesiw.io/moxie: Generate mocks with proxy functions

Go has a convenient feature called [embedding][embedding]. Method calls are
transparently passed through to embedded types.

`moxie` generates proxy methods that are only added during `go test`.

## Usage

Add this line to your code, where `T` is the type you want to instrument, and
`E` is the embedded interface whose functions you want to intercept.

``` go
type E interface {}

//go:generate go run lesiw.io/moxie@latest T
type T struct {
    E
}
```

Then run `go generate`.

## Functions

`moxie` makes the following methods on `T` available at test time.

### Instance methods

```go
(*T)._Func_Stub()      // when called, return zero values.
(*T)._Func_Return(...) // when called, return these values.

(*T)._Func_Do(func() {...}) // when called, run this instead.
                            // calling this with nil will un-mock the function.

(*T)._Func_Calls() []_T_Func_Call // return calls to Func.
```

Mocks queue. For example, mocking with `(*T)._Func_Return` twice will return the
first set of values the next time the function is called, then return the second
set of values in subsequent calls.

### Global methods

```go
new(T)._Func_StubAll(*testing.T)         // when called, return zero values.
new(T)._Func_ReturnAll(*testing.T, ...)  // when called, return these values.

new(T)._Func_DoAll(*testing.T, func() {...}) // when called, run this instead.
                                             // calling this with nil
                                             // will un-mock the function.

new(T)._Func_AllCalls() []_T_Func_Call // return calls to Func.
new(T)._Func_BubbleCalls(*testing.T)   // clear calls before and after test.
```

As above, mocks queue.

These methods may be easier to use in some cases, as they operate on _all_ `T`
types. This can be helpful in situations where injecting a mock value into the
code under test may be difficult to do. Note, however, that this makes them
unsafe for use with `t.Parallel()`.

Methods that accept a `*testing.T` will clean up mocks at the end of their
corresponding test or subtest.

Tests using `AllCalls()` should also call `new(T)._Func_BubbleCalls(t)`,
otherwise `AllCalls()` may also contain calls from other tests. 

[embedding]: https://go.dev/doc/effective_go#embedding

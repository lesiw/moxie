# Struct example (broken at the time of writing)

This example shows a broken case when embedding a struct. The generated file will contain a "testing" unused import and
the stub won't be populated with any methods.

## How to run this example

```
go generate && go test ./...`
```

package main

type Checker interface {
	Check() bool
}

//go:generate go run lesiw.io/moxie@latest Person
type Person struct {
	Checker
}

func main() {}

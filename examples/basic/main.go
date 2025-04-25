package main

type Checker interface {
	Check() bool
}

//go:generate go run lesiw.io/moxie@latest Person
type Person struct {
	Checker
}

func (p *Person) Check() bool {
	return true
}

func main() {}

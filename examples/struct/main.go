package main

type Alive struct {
	age int
}

func (a *Alive) IsAlive() bool {
	return a.age > 0
}

//go:generate go run lesiw.io/moxie@latest Person
type Person struct {
	Alive
}

func main() {}

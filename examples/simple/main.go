package main

type Worker interface {
	IsWorkingWell() bool
}

// Stub for testing Company
//
//go:generate go run lesiw.io/moxie@latest StubForWorker
type StubForWorker struct {
	Worker
}

// This is your actual struct that you want to test
type Company struct {
	Workers []Worker
}

func (c *Company) IsGonnaBeRich() bool {
	for _, worker := range c.Workers {
		if !worker.IsWorkingWell() {
			return false
		}
	}
	return true
}

func main() {}

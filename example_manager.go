package main

import (
	"github.com/spouk/go-experiments/goroutines"
)

func main() {
	m  := goroutines.NewManager()
	m.Run(10, 50, 6)
}

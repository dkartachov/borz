package main

import (
	"os"

	"github.com/dkartachov/borz/internal/worker"
)

func main() {
	worker.Run(os.Args[1:])
}

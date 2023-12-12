package main

import (
	"os"

	"github.com/dkartachov/borz/internal/borzlet"
)

func main() {
	borzlet.Run(os.Args[1:])
}

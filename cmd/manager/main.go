/*
The borzlet acts as the orchestration system's manager.
*/
package main

import (
	"os"

	"github.com/dkartachov/borz/internal/manager"
)

func main() {
	manager.Run(os.Args[1:])
}

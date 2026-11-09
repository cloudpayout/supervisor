package main

import (
	"fmt"
	"os"

	"github.com/cloudpayout/supervisor/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
		os.Exit(1)
	}
}

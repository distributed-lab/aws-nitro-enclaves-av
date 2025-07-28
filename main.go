package main

import (
	"os"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}

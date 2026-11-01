// chezmoi-agents manages and validates AI harness configurations and parity.
package main

import (
	"context"
	"os"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/cli"
)

func main() {
	code := cli.Run(context.Background(), os.Args[1:])
	os.Exit(code)
}

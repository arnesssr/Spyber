// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"os"

	"github.com/arnesssr/Spyber/internal/interface/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}

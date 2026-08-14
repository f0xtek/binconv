// Command binconv converts numbers between binary, octal, decimal and
// hexadecimal.
package main

import (
	"os"

	"github.com/f0xtek/binconv/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

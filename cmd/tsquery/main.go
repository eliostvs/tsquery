package main

import (
	"os"

	tsquery "github.com/eliostvs/tsquery/internal"
)

func main() {
	os.Exit(tsquery.CLI(os.Args[:], os.Stdout, os.Stderr))
}

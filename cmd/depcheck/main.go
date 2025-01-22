package main

import (
	"depcheck"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(depcheck.Analyzer) }

package main

import (
	"github.com/v-standard/go-depcheck"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(depcheck.Analyzer) }

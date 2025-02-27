package main

import (
	"flag"
	"x/pkg/modelgen"
)

func main() {
	var packagePath string
	var dialect string
	flag.StringVar(&packagePath, "package", ".", "path to the package to generate the query builder for")
	flag.StringVar(&dialect, "dialect", "mysql", "dialect to generate the query builder for")
	flag.Parse()

	modelgen.Generate(dialect, packagePath)
}

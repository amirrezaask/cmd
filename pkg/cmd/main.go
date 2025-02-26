package main

import (
	"flag"
)

func main() {
	var mode string
	var root string
	flag.StringVar(&mode, "mode", "all", "mode to run the generator in [config, application, cmd, controller, models, all]")

	flag.StringVar(&root, "root", ".", "root directory of your project.")

	flag.Parse()
	flag.Usage()

}

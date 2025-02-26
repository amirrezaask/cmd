package main

import "x/pkg/appspec"

func main() {
	app :=
		appspec.
			New("CMS").
			WithModels().
			WithControllers().
			WithConfig().
			WithCmd().
			WithConfigProvider("env")
	// WithConfigProvider("vault").
	// WithConfigProvider("consul").

	app.Build() // will generate necessary codes and then compile
}

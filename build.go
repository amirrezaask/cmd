package main

import "heaven/pkg/appspec"

func main() {
	app :=
		appspec.
			New("CMS").
			WithCmd().
			WithConfig().
			WithConfigProvider("env").
			// WithConfigProvider("vault").
			// WithConfigProvider("consul").
			WithModels().
			WithControllers()

	app.Build() // will generate necessary codes and then compile
}

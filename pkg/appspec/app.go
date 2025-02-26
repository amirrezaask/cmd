package appspec

import "os"

type AppSpec struct {
	name                  string
	modelsPackageName     string
	controllerPackageName string
	configPackageName     string
	cmdPackageName        string
	excludePackages       []string
	configProviders       []string
}

func New(name string) *AppSpec {
	return &AppSpec{
		name: name,
	}

}

func (a *AppSpec) WithCmd(packageName ...string) *AppSpec {
	if len(packageName) < 1 {
		packageName = append(packageName, "cmd")
	}
	a.cmdPackageName = packageName[0]
	return a
}

func (a *AppSpec) WithModels(packageName ...string) *AppSpec {
	if len(packageName) < 1 {
		packageName = append(packageName, "models")
	}
	a.modelsPackageName = packageName[0]
	return a
}

func (a *AppSpec) WithControllers(packageName ...string) *AppSpec {
	if len(packageName) < 1 {
		packageName = append(packageName, "controllers")
	}
	a.controllerPackageName = packageName[0]
	return a
}

func (a *AppSpec) WithConfig(configPackageName ...string) *AppSpec {
	if len(configPackageName) < 1 {
		configPackageName = append(configPackageName, "config")
	}
	a.configPackageName = configPackageName[0]
	return a
}

func (a *AppSpec) WithConfigProvider(provider string) *AppSpec {
	a.configProviders = append(a.configProviders, provider)
	return a
}

func (a *AppSpec) ExcludePackages(pkgs ...string) *AppSpec {
	a.excludePackages = append(a.excludePackages, pkgs...)
	return a
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func (a *AppSpec) generateModels() {
	panicIfError(os.MkdirAll("models", 0755))

}

func (a *AppSpec) generateControllers() {
	panicIfError(os.MkdirAll("controllers", 0755))
}

func (a *AppSpec) generateConfig() {
	panicIfError(os.MkdirAll("config", 0755))
}

func (a *AppSpec) generateCmd() {
	panicIfError(os.MkdirAll("cmd", 0755))
}
func (a *AppSpec) Build() {
	// bootstrap
	a.generateModels()
	a.generateControllers()
	a.generateConfig()
	a.generateCmd()

}

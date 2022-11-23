package enclaves

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"io/ioutil"
	"os"
)

const (
	dependenciesUrl = "https://docs.kurtosis.com/reference/starlark-reference/#dependencies"
)

type KurtosisMod struct {
	Module Module `yaml:"module"`
}

type Module struct {
	ModuleName string `yaml:"name"`
}

func parseKurtosisMod(kurtosisModFilepath string) (*KurtosisMod, error) {
	kurtosisModContents, err := ioutil.ReadFile(kurtosisModFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stacktrace.NewError("Couldn't find a '%v' in the root of the package at '%v'. Packages are expected to have a '%v' at root; have a look at '%v' for more", modFilename, kurtosisModFilepath, modFilename, dependenciesUrl)
		}
		return nil, stacktrace.Propagate(err, "An error occurred while reading the '%v' file at '%v'", modFilename, kurtosisModFilepath)
	}

	var kurtosisModule KurtosisMod
	err = yaml.Unmarshal(kurtosisModContents, &kurtosisModule)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while parsing the '%v' file at '%v'", modFilename, kurtosisModFilepath)
	}

	if kurtosisModule.Module.ModuleName == "" {
		return nil, stacktrace.NewError("Field module.name in %v needs to be set and cannot be empty", modFilename)
	}

	return &kurtosisModule, nil
}

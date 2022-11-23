package enclaves

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"io/ioutil"
)

type KurtosisYml struct {
	Module Module `yaml:"module"`
}

type Module struct {
	ModuleName string `yaml:"name"`
}

func parseKurtosisYml(kurtosisYmlFilepath string) (*KurtosisYml, error) {
	kurtosisYmlContents, err := ioutil.ReadFile(kurtosisYmlFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while reading the '%v' file at '%v'", kurtosisYmlFilename, kurtosisYmlFilepath)
	}

	var kurtosisYml KurtosisYml
	err = yaml.Unmarshal(kurtosisYmlContents, &kurtosisYml)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while parsing the '%v' file at '%v'", kurtosisYmlFilename, kurtosisYmlFilepath)
	}

	if kurtosisYml.Module.ModuleName == "" {
		return nil, stacktrace.NewError("Field module.name in %v needs to be set and cannot be empty", kurtosisYmlFilename)
	}

	return &kurtosisYml, nil
}

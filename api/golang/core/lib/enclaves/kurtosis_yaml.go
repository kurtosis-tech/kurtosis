package enclaves

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	packagesUrl = "https://docs.kurtosis.com/concepts-reference/packages"
)

// fields are public because it's needed for YAML decoding
type KurtosisYaml struct {
	PackageName string `yaml:"name"`
}

func ParseKurtosisYaml(kurtosisYamlFilepath string) (*KurtosisYaml, error) {
	kurtosisYamlContents, err := os.ReadFile(kurtosisYamlFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stacktrace.NewError("Couldn't find a '%v' in the root of the package at '%v'. Packages are expected to have a '%v' at root; have a look at '%v' for more", kurtosisYamlFilename, kurtosisYamlFilepath, kurtosisYamlFilename, packagesUrl)
		}
		return nil, stacktrace.Propagate(err, "An error occurred while reading the '%v' file at '%v'", kurtosisYamlFilename, kurtosisYamlFilepath)
	}

	var kurtosisYaml KurtosisYaml
	err = yaml.Unmarshal(kurtosisYamlContents, &kurtosisYaml)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while parsing the '%v' file at '%v'", kurtosisYamlFilename, kurtosisYamlFilepath)
	}

	if kurtosisYaml.PackageName == "" {
		return nil, stacktrace.NewError("Field 'name', which is the Starlark package's name, in %v needs to be set and cannot be empty", kurtosisYamlFilename)
	}

	return &kurtosisYaml, nil
}

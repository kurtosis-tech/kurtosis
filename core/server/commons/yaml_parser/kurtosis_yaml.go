package yaml_parser

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"io/ioutil"
)

var noPackageNameFound = ""

type KurtosisYaml struct {
	PackageName string `yaml:"name"`
}

func (parser *KurtosisYaml) GetPackageName() string {
	if parser == nil {
		return noPackageNameFound
	}
	return parser.PackageName
}

// TODO: this parsing logic is similar to what have we in the api, maybe we should move everything into one
// common package. This method assumes that the kurtosis.yml exists in the path provided.
func parseKurtosisYamlInternal(absPathToKurtosisYaml string, read func(filename string) ([]byte, error)) (*KurtosisYaml, error) {
	kurtosisYamlContent, err := read(absPathToKurtosisYaml)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while reading the contents of '%v'", absPathToKurtosisYaml)
	}

	var kurtosisYaml KurtosisYaml
	if err = yaml.Unmarshal(kurtosisYamlContent, &kurtosisYaml); err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while analyzing the contents of '%v'", absPathToKurtosisYaml)
	}
	return &kurtosisYaml, nil
}

func ParseKurtosisYaml(absPathToKurtosisYaml string) (*KurtosisYaml, error) {
	return parseKurtosisYamlInternal(absPathToKurtosisYaml, ioutil.ReadFile)
}

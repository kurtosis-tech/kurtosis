package yaml_parser

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

var noPackageNameFound = ""
var noReplaceDependencies = map[string]string{}

type KurtosisYaml struct {
	PackageName         string            `yaml:"name"`
	ReplaceDependencies map[string]string `yaml:"replace"`
}

func (parser *KurtosisYaml) GetPackageName() string {
	if parser == nil {
		return noPackageNameFound
	}
	return parser.PackageName
}

func (parser *KurtosisYaml) GetReplaceDependencies() map[string]string {
	if parser == nil {
		return noReplaceDependencies
	}
	return parser.ReplaceDependencies
}

// TODO: this parsing logic is similar to what have we in the api, maybe we should move everything into one
// common package. This method assumes that the kurtosis.yml exists in the path provided.
func parseKurtosisYamlInternal(absPathToKurtosisYaml string, read func(filename string) ([]byte, error)) (*KurtosisYaml, error) {
	kurtosisYamlContent, err := read(absPathToKurtosisYaml)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while reading the contents of '%v'", absPathToKurtosisYaml)
	}

	var kurtosisYaml KurtosisYaml
	if err = yaml.UnmarshalStrict(kurtosisYamlContent, &kurtosisYaml); err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while analyzing the contents of '%v'", absPathToKurtosisYaml)
	}
	logrus.Debugf("parsed kurtosis.yml '%+v'", kurtosisYaml)
	return &kurtosisYaml, nil
}

func ParseKurtosisYaml(absPathToKurtosisYaml string) (*KurtosisYaml, error) {
	return parseKurtosisYamlInternal(absPathToKurtosisYaml, os.ReadFile)
}

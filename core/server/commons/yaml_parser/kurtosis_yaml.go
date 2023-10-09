package yaml_parser

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

var noPackageNameFound = ""
var naPackageDescriptionFound = ""
var noPackageReplaceOptions = map[string]string{}

type KurtosisYaml struct {
	PackageName           string            `yaml:"name"`
	PackageDescription    string            `yaml:"description"`
	PackageReplaceOptions map[string]string `yaml:"replace"`
}

func (parser *KurtosisYaml) GetPackageName() string {
	if parser == nil {
		return noPackageNameFound
	}
	return parser.PackageName
}

func (parser *KurtosisYaml) GetPackageDescription() string {
	if parser == nil {
		return naPackageDescriptionFound
	}
	return parser.PackageDescription
}

func (parser *KurtosisYaml) GetPackageReplaceOptions() map[string]string {
	if parser == nil {
		return noPackageReplaceOptions
	}
	return parser.PackageReplaceOptions
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

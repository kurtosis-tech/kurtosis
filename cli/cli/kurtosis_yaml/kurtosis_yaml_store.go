package kurtosis_yaml

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	kurtosisYmlFilename                     = "kurtosis.yml"
	kurtosisYamlFilePermissions os.FileMode = 0644
)

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func initializeKurtosisPackage(packageDirpath string, packageName string) error {

	//TODO validate the package name it's using github and org use the git package content provider methods

	kurtosisYaml := enclaves.KurtosisYaml{
		PackageName:           packageName,
		PackageDescription:    fmt.Sprintf("# %s\nEnter description Markdown here.", packageName), //TODO hardcoded values
		PackageReplaceOptions: map[string]string{},
	}

	kurtosisYAMLContent, err := yaml.Marshal(kurtosisYaml)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis yaml file '%+v'", kurtosisYaml)
	}

	logrus.Debugf("Creating the Kurtosis package YAML file...")

	if err := createKurtosisYamlFile(packageDirpath, kurtosisYAMLContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the '%s' on '%s'", kurtosisYmlFilename, packageDirpath)
	}

	logrus.Debugf("...Kurtosis package YAML file saved")
	return nil
}

func createKurtosisYamlFile(packageDirpath string, content []byte) error {
	kurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)

	fileInfo, err := os.Stat(kurtosisYamlFilepath)
	if fileInfo != nil {
		return stacktrace.NewError("Imposible to create a new Kurtosis package inside '%s' because a file with name '%s' already exist on this path", packageDirpath, kurtosisYmlFilename)
	}
	if os.IsNotExist(err) {
		err = os.WriteFile(kurtosisYamlFilepath, content, kurtosisYamlFilePermissions)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the Kurtosis package YAML file '%s'", kurtosisYamlFilepath)
		}
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the '%s' file on '%s'", kurtosisYmlFilename, packageDirpath)
	}
	return nil
}

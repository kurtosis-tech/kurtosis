package kurtosis_package

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/shared_utils"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	kurtosisPackageFilePermissions os.FileMode = 0644
	kurtosisYmlFilename                        = "kurtosis.yml"
	mainStarFilename                           = "main.star"
	mainStarFileContentStr                     = `def run(plan):
    # TODO
    plan.print("hello world!")`
	kurtosisYmlDescriptionFormat = "# %s\nEnter description Markdown here."
)

func InitializeKurtosisPackage(packageDirpath string, packageName string, isExecutablePackage bool) error {

	// validate package name
	_, err := shared_utils.ParseGitURL(packageName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred validating package name '%v', invalid GitHub URL", packageName)
	}

	logrus.Debugf("Initializaing the '%s' Kurtosis package...", packageName)
	if err := createKurtosisYamlFile(packageDirpath, packageName); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the '%s' on '%s'", kurtosisYmlFilename, packageDirpath)
	}

	if isExecutablePackage {
		if err := createMainStarFile(packageDirpath); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the '%s' on '%s'", mainStarFilename, packageDirpath)
		}
	}

	logrus.Debugf("...'%s' Kurtosis package successfully initialized", packageName)

	return nil
}

func createKurtosisYamlFile(packageDirpath string, packageName string) error {

	defaultPackageDescriptionForInitPackage := fmt.Sprintf(kurtosisYmlDescriptionFormat, packageName)
	defaultPackageReplaceOptionsForInitPackage := map[string]string{}

	kurtosisYaml := enclaves.NewKurtosisYaml(packageName, defaultPackageDescriptionForInitPackage, defaultPackageReplaceOptionsForInitPackage)

	kurtosisYAMLContent, err := yaml.Marshal(kurtosisYaml)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis yaml file '%+v'", kurtosisYaml)
	}

	kurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)

	fileInfo, err := os.Stat(kurtosisYamlFilepath)
	if fileInfo != nil {
		return stacktrace.NewError("Imposible to create a new Kurtosis package inside '%s' because a file with name '%s' already exist on this path", packageDirpath, kurtosisYmlFilename)
	}
	if os.IsNotExist(err) {
		logrus.Debugf("Creating the '%s' file...", kurtosisYmlFilename)
		err = os.WriteFile(kurtosisYamlFilepath, kurtosisYAMLContent, kurtosisPackageFilePermissions)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the '%s' file", kurtosisYamlFilepath)
		}
		logrus.Debugf("...'%s' file created", kurtosisYmlFilename)
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the '%s' file on '%s'", kurtosisYmlFilename, packageDirpath)
	}

	return nil
}

func createMainStarFile(packageDirpath string) error {

	mainStarFilepath := path.Join(packageDirpath, mainStarFilename)

	fileInfo, err := os.Stat(mainStarFilepath)
	if fileInfo != nil {
		return stacktrace.NewError("Imposible to create a new Kurtosis package inside '%s' because a file with name '%s' already exist on this path", packageDirpath, mainStarFilename)
	}

	mainStarFileContent := []byte(mainStarFileContentStr)

	if os.IsNotExist(err) {
		logrus.Debugf("Creating the '%s' file...", mainStarFilename)
		err = os.WriteFile(mainStarFilepath, mainStarFileContent, kurtosisPackageFilePermissions)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the '%s' file", mainStarFilepath)
		}
		logrus.Debugf("...'%s' file created", mainStarFilename)
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the '%s' file on '%s'", mainStarFilename, packageDirpath)
	}
	return nil
}

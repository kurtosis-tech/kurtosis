package main

import (
	"fmt"
	positional_arg_parser "github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/commons/repl_consts"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
	"text/template"
)

const (
	binaryFilepathArg = "binary-filepath"
	dockerfileTemplateFilepathArg = "dockerfile-template"
	outputFilepathArg = "output-filepath"
	replTypeArg = "repl-type"
)
var positionalArgs = []string{
	binaryFilepathArg,
	dockerfileTemplateFilepathArg,
	outputFilepathArg,
	replTypeArg,
}


type TemplateData struct {
	KurtosisCoreVersion           string
	PackageInstallationDirpath    string
	InstalledPackagesDirpath      string
	KurtosisAPIContainerIPEnvVar  string
	KurtosisAPIContainerPortEnvVar string
	EnclaveIDEnvVar               string
	EnclaveDataMountDirpathEnvVar string
}

// This is a tiny Go script that uses the Kurtosis client version exposed in Go code to generate
//  REPL Docker images, to ensure that the REPL images are always using the right version of Kurt Client
//  and we don't have to remember to update it manually
func main() {
	if err := runMain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}




func runMain() error {
	parsedArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, os.Args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	templateFilepath := parsedArgs[dockerfileTemplateFilepathArg]
	outputFilepath := parsedArgs[outputFilepathArg]
	replType := repl_consts.ReplType(parsedArgs[replTypeArg])

	// For some reason, the template name has to match the basename of the file:
	//  https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
	templateFilename := path.Base(templateFilepath)
	tmpl, err := template.New(templateFilename).ParseFiles(templateFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the template '%v'", templateFilepath)
	}

	packageInstallationDirpath, found := repl_consts.PackageInstallationDirpaths[replType]
	if !found {
		return stacktrace.NewError("No package installation dirpath defined for REPL type '%v'", replType)
	}
	installedPackagesDirpath, found := repl_consts.InstalledPackagesDirpath[replType]
	if !found {
		return stacktrace.NewError("No installed packages dirpath defined for REPL type '%v'", replType)
	}


	data := TemplateData{
		KurtosisCoreVersion:           api_container_launcher.DefaultVersion,
		PackageInstallationDirpath:    packageInstallationDirpath,
		InstalledPackagesDirpath:      installedPackagesDirpath,
		KurtosisAPIContainerIPEnvVar: repl_consts.KurtosisAPIContainerIPEnvVar,
		KurtosisAPIContainerPortEnvVar: repl_consts.KurtosisAPIContainerPortEnvVar,
		EnclaveIDEnvVar:               repl_consts.EnclaveIdEnvVar,
		EnclaveDataMountDirpathEnvVar: repl_consts.EnclaveDataMountDirpathEnvVar,
	}

	fp, err := os.Create(outputFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening the output file for writing")
	}
	defer fp.Close()

	if err := tmpl.Execute(fp, data); err != nil {
		return stacktrace.Propagate(err, "An error occurred filling the template")
	}
	return nil
}

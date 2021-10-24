package main

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/kurtosis_api_version_const"
	"github.com/palantir/stacktrace"
	"os"
	"path"
	"text/template"
)

type TemplateData struct {
	KurtosisClientVersion string
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
	// 3, because arg 0 is the filepath of the binary
	if len(os.Args) != 3 {
		return stacktrace.NewError("Expected exactly two args 1) path to the Dockerfile template 2) output filepath")
	}
	templateFilepath := os.Args[1]
	outputFilepath := os.Args[2]

	// For some reason, the template name has to match teh basename of the file:
	//  https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
	templateFilename := path.Base(templateFilepath)
	tmpl, err := template.New(templateFilename).ParseFiles(templateFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the template '%v'", templateFilepath)
	}

	data := TemplateData{
		KurtosisClientVersion: kurtosis_api_version_const.KurtosisApiVersion,
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

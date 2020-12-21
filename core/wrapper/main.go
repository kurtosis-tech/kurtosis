/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"text/template"
)

const (
	// "Enum" of actions
	StoreTrue Action =  iota
	StoreValue

	flagPrefix = "--"

	failureExitCode = 1
)

type Action int

type TemplateData struct {
	OneLinerHelpText string
	LinewiseHelpText []string
}

type Argument struct {
	// If empty, this is a positional Argument
	Flag string

	// The Bash Variable that the value will be stored to
	Variable string

	// The default Variable that the Bash Variable will be assigned to, if not present
	DefaultVal string

	// The Variable HelpText
	HelpText string

	// The Action taken (only relevant for Flag variables; position variables will always store the value)
	Action Action
}

var args = []Argument{
	{
		Flag:       "--parallelism",
		Variable:   "parallelism",
		DefaultVal: "4",
		HelpText:   "The number of texts to execute in parallel",
		Action:     StoreValue,
	},
	{
		Flag:       "--help",
		Variable:   "show_help",
		DefaultVal: "false",
		HelpText:   "Display this message",
		Action:     StoreTrue,
	},
}

// Fills the Bash wrapper script template with the appropriate variables
func main()  {
	templateFilepathArg := flag.String(
		"template",
		"",
		"Filepath containing Bash template file that will get rendered into the Kurtosis wrapper script",
	)
	/*
	outputFilepathArg := flag.String(
		"output",
		"",
		"Output filepath to write the rendered template to",
	)

	 */
	flag.Parse()

	tmpl, err := template.New("wrapper").ParseFiles(*templateFilepathArg)
	if err != nil {
		logrus.Errorf("An error occurred parsing the Bash template: %v", err)
		os.Exit(failureExitCode)
	}

	data, err := GenerateTemplateData(args)
	if err != nil {
		logrus.Errorf("An error occurred generating the template data: %v", err)
		os.Exit(failureExitCode)
	}

	if err := tmpl.Execute(os.Stdout, data); err != nil {
		logrus.Errorf("An error occurred filling the template: %v", err)
		os.Exit(failureExitCode)
	}
}

func GenerateTemplateData(args []Argument) (*TemplateData, error) {
	// Verify that all flag args start with the appropriate prefix

	flagArgsOnelinerFragments := []string{}
	positionalArgsOnelinerFragments := []string{}

	flagArgsLinewiseHelptext := []string{}
	positionalArgsLinewiseHelptext := []string{}
	for _, arg := range args {
		isFlagArg := arg.Flag != ""

		if isFlagArg {
			if !strings.HasPrefix(arg.Flag, flagPrefix) {
				return nil, stacktrace.NewError(
					"Flag '%v' must start with flag prefix '%v'",
					arg.Flag,
					flagPrefix)
			}

			var onelinerText string
			if (arg.Action == StoreValue) {
				onelinerText = fmt.Sprintf("%v %v", arg.Flag, arg.Variable)
			} else if (arg.Action == StoreTrue) {
				onelinerText = arg.Flag
			} else {
				return nil, stacktrace.NewError("Unrecognized arg Action '%v'; this is a code bug", arg.Action)
			}
			flagArgsOnelinerFragments = append(flagArgsOnelinerFragments, onelinerText)

			var linewiseText string
			if (arg.DefaultVal != "") {
				linewiseText = fmt.Sprintf(
					"%v\t%v (default: %v)",
					onelinerText,
					arg.HelpText,
					arg.DefaultVal,
				)
			} else {
				linewiseText = fmt.Sprintf(
					"%v\t%v",
					onelinerText,
					arg.HelpText,
				)
			}
			flagArgsLinewiseHelptext = append(flagArgsLinewiseHelptext, linewiseText)
		} else {
			positionalArgsOnelinerFragments = append(positionalArgsOnelinerFragments, arg.Variable)

			linewiseText := fmt.Sprintf(
				"%v\t%v",
				arg.Variable,
				arg.HelpText,
			)
			positionalArgsLinewiseHelptext = append(positionalArgsLinewiseHelptext, linewiseText)
		}
	}
	flagArgsOneliner := strings.Join(flagArgsOnelinerFragments, " ")
	positionalArgsOneliner := strings.Join(positionalArgsOnelinerFragments, " ")
	combinedOneliner := fmt.Sprintf("%v %v", flagArgsOneliner, positionalArgsOneliner)

	combinedLinewiseHelptext := append(flagArgsLinewiseHelptext, positionalArgsLinewiseHelptext...)
	return &TemplateData{
		OneLinerHelpText: combinedOneliner,
		LinewiseHelpText: combinedLinewiseHelptext,
	}, nil
}




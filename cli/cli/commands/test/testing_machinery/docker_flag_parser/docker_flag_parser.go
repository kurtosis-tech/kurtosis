/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_flag_parser

/*
Because the initializer is run via Docker, the arguments that a user needs to provide are NOT the same as the CLI
flags that get passed in to main.go (e.g. to set parallelism, main.go might ask for a "--parallelism=X" argument
but the user really sets parallelism by passing in a Docker environment variable of "--env PARALLELISM=X"). As
such, the default fmt.Usage function can't be used because it will print things that are confusing to a user.

To fix this, we override the flag-setting and usage-printing with our own custom logic, so that helptext is
actually instructive to the user.
 */

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	requiredArgDescription = "required"
	optionalArgDescription = "optional"

	// "Enum" for flag types
	BoolFlagType   FlagType = "bool"
	StringFlagType FlagType = "string"
	IntFlagType    FlagType = "int"
)

type FlagType string

type FlagConfig struct {
	Required bool
	HelpText string
	Default  interface{}
	Type     FlagType
}

type FlagParser struct {
	flagConfigs map[string]FlagConfig
}

func NewFlagParser(flagConfigs map[string]FlagConfig) *FlagParser {
	return &FlagParser{
		flagConfigs: flagConfigs,
	}
}

func (parser *FlagParser) Parse() (*ParsedFlags, error) {
	boolFlags := map[string]*bool{}
	stringFlags := map[string]*string{}
	intFlags := map[string]*int{}

	for name, config := range parser.flagConfigs {
		var parseFunc func()
		switch config.Type {
		case BoolFlagType:
			boolDefaultValue := config.Default.(bool)
			parseFunc = func() {
				// Blank helptext to emphasize that we do our own usage function
				boolFlags[name] = flag.Bool(name, boolDefaultValue, "")
			}
		case StringFlagType:
			stringDefaultValue := config.Default.(string)
			parseFunc = func() {
				// Blank helptext to emphasize that we do our own usage function
				stringFlags[name] = flag.String(name, stringDefaultValue, "")
			}
		case IntFlagType:
			intDefaultValue := config.Default.(int)
			parseFunc = func() {
				// Blank helptext to emphasize that we do our own usage function
				intFlags[name] = flag.Int(name, intDefaultValue, "")
			}
		default:
			return nil, stacktrace.NewError("Unrecognized flag type '%v'", config.Type)
		}
		parseFunc()
	}

	flag.Usage = func() {
		parser.ShowUsage()
	}
	flag.Parse()

	return &ParsedFlags{
		boolFlags:   boolFlags,
		stringFlags: stringFlags,
		intFlags: intFlags,
	}, nil
}

func (parser *FlagParser) ShowUsage() error {
	sortedFlagNames := []string{}
	for name, _ := range parser.flagConfigs {
		sortedFlagNames = append(sortedFlagNames, name)
	}
	sort.Strings(sortedFlagNames)

	fmt.Println("Docker environment variable 'arguments' that can be passed in using '--env ARGNAME=ARGVALUE':")
	fmt.Println()
	for _, name := range sortedFlagNames {
		config := parser.flagConfigs[name]

		var defaultValueSuffix string
		switch config.Type {
		case BoolFlagType:
			boolDefaultValue := config.Default.(bool)
			defaultValueSuffix = fmt.Sprintf(" (default: %v)", boolDefaultValue)
		case StringFlagType:
			stringDefaultValue := config.Default.(string)
			if stringDefaultValue == "" {
				defaultValueSuffix = ""
			} else {
				defaultValueSuffix = fmt.Sprintf(" (default: %v)", stringDefaultValue)
			}
		case IntFlagType:
			intDefaultValue := config.Default.(int)
			defaultValueSuffix = fmt.Sprintf(" (default: %v)", intDefaultValue)
		default:
			return stacktrace.NewError("Unrecognized flag type '%v'", config.Type)
		}

		var requiredOrOptional string
		if config.Required {
			requiredOrOptional = requiredArgDescription
		} else {
			requiredOrOptional = optionalArgDescription
		}

		fmt.Printf("   %v (%v %v):\n", name, requiredOrOptional, config.Type)
		fmt.Printf("      %v%v\n", config.HelpText, defaultValueSuffix)
		fmt.Println()
	}
	fmt.Println("IMPORTANT: Docker doesn't like unescaped spaces when using the '--env' flag, so make sure you backslash-escape spaces in your environment variable values like so: --env TEST_NAMES=\"my\\ specific\\ test1\"")
	fmt.Println()
	return nil
}

// ============================= PARSED FLAGS =======================================
type ParsedFlags struct {
	boolFlags map[string]*bool
	stringFlags map[string]*string
	intFlags map[string]*int
}

func (parsed *ParsedFlags) GetBool(name string) bool {
	return *parsed.boolFlags[name]
}

func (parsed *ParsedFlags) GetString(name string) string {
	return *parsed.stringFlags[name]
}

func (parsed *ParsedFlags) GetInt(name string) int {
	return *parsed.intFlags[name]
}

package kurtosis_command

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/pflag"
)

// Struct-based enum: https://threedots.tech/post/safer-enums-in-go/
type FlagType struct {
	// Private so users can't instantiate it - they have to use our enum values
	typeStr string
}
var (
	FlagType_Uint32 = FlagType{typeStr: "uint32"}
	FlagType_String = FlagType{typeStr: "string"}
	FlagType_Bool = FlagType{typeStr: "bool"}
)

// TODO Maybe better to make several different types of flags here - one for each type of value
type FlagConfig struct {
	// Long-form name
	Key string

	// A single-character shorthand for the flag
	Shorthand byte

	Type FlagType

	Default string

	// TODO Use this!
	ValidationFunc func(string) error
}

type ParsedFlags struct {
	cmdFlagsSet *pflag.FlagSet
}
func (flags *ParsedFlags) GetString(name string) (string, error) {
	value, err := flags.cmdFlagsSet.GetString(name)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting string flag '%v'",
			name,
		)
	}
	return value, nil
}
func (flags *ParsedFlags) GetUint32(name string) (uint32, error) {
	value, err := flags.cmdFlagsSet.GetUint32(name)
	if err != nil {
		return 0, stacktrace.Propagate(err,
			"An error occurred getting uint32 flag '%v'",
			name,
		)
	}
	return value, nil
}
func (flags *ParsedFlags) GetBool(name string) (bool, error) {
	value, err := flags.cmdFlagsSet.GetBool(name)
	if err != nil {
		return 0, stacktrace.Propagate(err,
			"An error occurred getting bool flag '%v'",
			name,
		)
	}
	return value, nil
}

package flags

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/pflag"
)

type ParsedFlags struct {
	cmdFlagsSet *pflag.FlagSet
}

func NewParsedFlags(cmdFlagsSet *pflag.FlagSet) *ParsedFlags {
	return &ParsedFlags{cmdFlagsSet: cmdFlagsSet}
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
		return false, stacktrace.Propagate(err,
			"An error occurred getting bool flag '%v'",
			name,
		)
	}
	return value, nil
}

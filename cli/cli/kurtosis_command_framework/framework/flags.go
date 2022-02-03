package framework

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/pflag"
)

// TODO Maybe better to make several different types of flags here - one for each type of value
type FlagConfig struct {
	Key string

	// TODO Add flag defaults!!!

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
			"An error occurred getting string flag '%v'; this is a bug in Kurtosis!",
			name,
		)
	}
	return value, nil
}
func (flags *ParsedFlags) GetUint32(name string) (uint32, error) {
	value, err := flags.cmdFlagsSet.GetUint32(name)
	if err != nil {
		return 0, stacktrace.Propagate(err,
			"An error occurred getting uint32 flag '%v'; this is a bug in Kurtosis!",
			name,
		)
	}
	return value, nil
}

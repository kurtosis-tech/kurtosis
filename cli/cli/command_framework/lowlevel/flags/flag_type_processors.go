package flags

import (
	"github.com/kurtosis-tech/stacktrace"
	flag "github.com/spf13/pflag"
	"strconv"
)

const (
	uintBase = 10
	uint32Bits = 32
)

// flagTypeProcessor will take in a user's requested default value for the flag (serialized as a string),
// downcast the default value to the appropriate type, and attach it to the given Cobra flags set
type flagTypeProcessor func(
	flagKey string,
	shorthand string,
	defaultValueStr string,
	usage string,
	cobraFlagSet *flag.FlagSet,
) error

// Completeness enforced via unit test
var AllFlagTypeProcessors = map[FlagType]flagTypeProcessor{
	FlagType_String: processStringFlag,
	FlagType_Uint32: processUint32Flag,
	FlagType_Bool: processBoolFlag,
}

func processStringFlag(
	flagKey string,
	shorthand string,
	defaultValueStr string,
	usage string,
	cobraFlagSet *flag.FlagSet,
) error {
	// No validation needed because the default type is already string
	cobraFlagSet.StringP(
		flagKey,
		shorthand,
		defaultValueStr,
		usage,
	)
	return nil
}


func processUint32Flag(
	flagKey string,
	shorthand string,
	defaultValueStr string,
	usage string,
	cobraFlagSet *flag.FlagSet,
) error {
	defaultValueUint64, err := strconv.ParseUint(defaultValueStr, uintBase, uint32Bits)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"Could not parse default value string '%v' of flag '%v' to a uint32 using base %v and bits %v",
			defaultValueStr,
			flagKey,
			uintBase,
			uint32Bits,
		)
	}
	cobraFlagSet.Uint32P(
		flagKey,
		shorthand,
		uint32(defaultValueUint64),
		usage,
	)
	return nil
}

func processBoolFlag(
	flagKey string,
	shorthand string,
	defaultValueStr string,
	usage string,
	cobraFlagSet *flag.FlagSet,
) error {
	defaultValue, err := strconv.ParseBool(defaultValueStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"Could not parse default value string '%v' of flag '%v' to a bool",
			defaultValueStr,
			flagKey,
		)
	}
	cobraFlagSet.BoolP(
		flagKey,
		shorthand,
		defaultValue,
		usage,
	)
	return nil
}

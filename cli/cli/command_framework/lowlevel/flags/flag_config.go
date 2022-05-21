package flags

type FlagConfig struct {
	// Long-form name of the flag
	Key string

	// TODO Rename this to "description"
	// Usage string
	Usage string

	// A single-character shorthand for the flag
	// If shorthand is emptystring, no shorthand will be used
	Shorthand string

	// Used for validating the flag
	// If not set, this will default to FlagType_String
	Type FlagType

	// Default, serialized as a string, that will be displayed in the usage
	Default string

	// TODO Validation function

	// TODO Add the ability to have greedy params!!! Would be very useful for things like '--ports' and '--files' in 'service add'
}
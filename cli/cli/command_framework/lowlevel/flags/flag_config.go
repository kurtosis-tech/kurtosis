package flags

const (
	uint32Base = 10
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
func (flagType *FlagType) AsString() string {
	return flagType.typeStr
}

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
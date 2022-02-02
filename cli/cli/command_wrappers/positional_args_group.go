package command_wrappers

import "github.com/kurtosis-tech/stacktrace"

// Cases:
// Simple required: enclave-id
// Complex required: enclave-id service-id
// Simple optional: [enclave-id]
//
// Complex optional: [enclave-id service-id]
// N>=0: [enclave-id...]
// N>=1: enclave-id [enclave-id...]
//	NOTE: This can actually be done via two args!

// This represents one or more positional arguments
type PositionalArgsGroup interface {
	getArgStrConsumer() PositionalArgStrConsumer

	// TODO something about the expected number of arguments?

	// TODO something about optional arguments? (though maybe this can be rolled into the number of arguments consumed)

	// The value extractor is a function that will run over the input strings that the user passes in, run any required
	//  validation, and return a map of the
	extractValues(strs []string) (map[PositionalArgKey]string, error)
}

// A positional args group where the
type simplePositionalArgsGroup struct {
	// The key that will be displayed for each arg (e.g. for enclave ID, this could be "enclave-id")
	argKey string
}
// TODO constructor
func (simple *simplePositionalArgsGroup) extractValues(strs []string) (map[PositionalArgKey]string, error) {
	if len(strs) != 1 {
		return nil, stacktrace.NewError("Expected ")
	}
}

type

// A positional args group composed of multiple other positional args groups
type compositePositionalArgsGroup struct {

}
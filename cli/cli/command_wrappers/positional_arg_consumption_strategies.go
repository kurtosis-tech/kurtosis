package command_wrappers

type PositionalArgStrConsumer interface {
	// From a list of remaining positional arg strs that need to be consumed, spit out the number of arg strings that the consumer wants to consume
	GetDesiredNumArgsConsumed(input []string) uint32
}

// Consumes a fixed number of positional args
type fixedNumberPositionalArgConsumer struct {
	numArgsToConsume uint32
}
func (fixed *fixedNumberPositionalArgConsumer) GetDesiredNumArgsConsumed(input []string) uint32 {
	return fixed.numArgsToConsume
}

// Consumes as many arg strings as possible
type unboundedPositionalArgConsumer struct {}
func (unbounded *unboundedPositionalArgConsumer) GetDesiredNumArgsConsumed(input []string) uint32 {
	return uint32(len(input))
}



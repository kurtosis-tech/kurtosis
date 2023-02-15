package flags

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlagTypeProcessorsCompleteness(t *testing.T) {
	for _, flagType := range FlagTypeValues() {
		_, found := AllFlagTypeProcessors[flagType]
		require.True(t, found, "Missing a flag type processor for flag type '%v'", flagType)
	}
}

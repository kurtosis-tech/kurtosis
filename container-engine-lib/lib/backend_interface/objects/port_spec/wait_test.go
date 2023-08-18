package port_spec

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWaitMarshallers(t *testing.T) {
	originalWait := NewWait(4 * time.Minute)

	marshaledWait, err := json.Marshal(originalWait)
	require.NoError(t, err)
	require.NotNil(t, marshaledWait)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newWait := &Wait{}

	err = json.Unmarshal(marshaledWait, newWait)
	require.NoError(t, err)

	require.EqualValues(t, originalWait, newWait)
}

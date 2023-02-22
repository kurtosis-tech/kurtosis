package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getAdjectives_orderShouldRemainSame(t *testing.T) {
	firstCall := getAdjectives()
	secondCall := getAdjectives()
	require.Equal(t, firstCall, secondCall)
}

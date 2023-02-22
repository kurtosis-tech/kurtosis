package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNouns_orderShouldRemainSame(t *testing.T) {
	firstCall := getNouns()
	secondCall := getNouns()

	require.Equal(t, firstCall, secondCall)
}

package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomNameGenerator_GetNameGenerator(t *testing.T) {
	nameGenerator := getNameGenerator()
	nameGenerator2 := getNameGenerator()
	// testing the memory address equality
	require.True(t, nameGenerator == nameGenerator2)
}

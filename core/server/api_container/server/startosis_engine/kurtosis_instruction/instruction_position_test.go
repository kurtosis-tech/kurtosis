package kurtosis_instruction

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestInstructionPosition_RegularExpressionAndPlaceholderAlign(t *testing.T) {
	dummySuffix := "dummySuffix"
	placeHolderStr := fmt.Sprintf(placeholderFormat, 5, 6, dummySuffix)
	compiledRegex := regexp.MustCompile(fmt.Sprintf(regexFormat, dummySuffix))
	hasMatches := compiledRegex.MatchString(placeHolderStr)
	require.True(t, hasMatches)
}

package kurtosis_instruction

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestInstructionPosition_RegularExpressionAndPlaceholderAlign(t *testing.T) {
	dummySuffix := "dummySuffix"
	placeHolderStr := fmt.Sprintf(magicStringFormat, "github.com/kurtosis-tech/eth2-module/src/participant_network/prelaunch_data_generator/el_genesis/el_genesis_data_generator.star", 5, 6, dummySuffix)
	compiledRegex := regexp.MustCompile(fmt.Sprintf(regexFormat, dummySuffix))
	hasMatches := compiledRegex.MatchString(placeHolderStr)
	require.True(t, hasMatches)
}

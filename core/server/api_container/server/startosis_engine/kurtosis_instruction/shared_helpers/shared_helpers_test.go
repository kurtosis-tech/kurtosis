package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceId = "test-service-id"
)

func TestReplaceMagicStringWithValue_SimpleCase(t *testing.T) {
	instruction := kurtosis_instruction.NewInstructionPosition(5, 3)
	inputStr := instruction.MagicString(ArtifactUUIDSuffix)
	environment := startosis_executor.NewExecutionEnvironment()
	testUuid := "test-uuid"
	environment.SetArtifactUuid(inputStr, testUuid)

	expectedOutput := testUuid
	replacedStr, err := ReplaceMagicStringWithValue(ArtifactUUIDSuffix, inputStr, testServiceId, environment)

	require.Nil(t, err)
	require.Equal(t, expectedOutput, replacedStr)
}

func TestReplaceMagicStringWithValue_ValidMultipleReplaces(t *testing.T) {
	instructionA := kurtosis_instruction.NewInstructionPosition(5, 3)
	instructionB := kurtosis_instruction.NewInstructionPosition(56, 33)
	magicStringA := instructionA.MagicString(ArtifactUUIDSuffix)
	magicStringB := instructionB.MagicString(ArtifactUUIDSuffix)
	inputStr := fmt.Sprintf("%v %v %v", magicStringB, magicStringA, magicStringB)

	environment := startosis_executor.NewExecutionEnvironment()
	testUuidA := "test-uuid-a"
	testUuidB := "test-uuid-b"
	expectedOutput := fmt.Sprintf("%v %v %v", testUuidB, testUuidA, testUuidB)

	environment.SetArtifactUuid(magicStringA, testUuidA)
	environment.SetArtifactUuid(magicStringB, testUuidB)

	replacedStr, err := ReplaceMagicStringWithValue(ArtifactUUIDSuffix, inputStr, testServiceId, environment)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, replacedStr)
}

func TestReplaceMagicStringWithValue_MagicStringNotInEnvironment(t *testing.T) {
	instruction := kurtosis_instruction.NewInstructionPosition(5, 3)
	magicString := instruction.MagicString(ArtifactUUIDSuffix)
	emptyEnvironment := startosis_executor.NewExecutionEnvironment()
	_, err := ReplaceMagicStringWithValue(ArtifactUUIDSuffix, magicString, testServiceId, emptyEnvironment)
	require.NotNil(t, err)
}

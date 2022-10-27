package startosis_errors

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInterpretationError_serializationSimpleError(t *testing.T) {
	errorString := "This is an error!"
	errorToSerialize := NewInterpretationError(errorString)

	require.Equal(t, errorString, errorToSerialize.Error())
}

func TestInterpretationError_serializationWithCustomMsg(t *testing.T) {
	errorToSerialize := NewInterpretationErrorWithCustomMsg(
		[]CallFrame{
			*NewCallFrame("<toplevel>", NewScriptPosition(13, 12)),
			*NewCallFrame("add_service", NewScriptPosition(0, 0)),
		},
		"Evaluation error: Missing `container_image_name` as part of the struct object",
	)

	expectedOutput := `Evaluation error: Missing ` + "`container_image_name`" + ` as part of the struct object
	at [13:12]: <toplevel>
	at [0:0]: add_service`
	require.Equal(t, expectedOutput, errorToSerialize.Error())
}

func TestInterpretationError_serializationFromStacktrace(t *testing.T) {
	errorToSerialize := NewInterpretationErrorFromStacktrace(
		[]CallFrame{
			*NewCallFrame("<toplevel>", NewScriptPosition(13, 12)),
			*NewCallFrame("add_service", NewScriptPosition(0, 0)),
		},
	)

	expectedOutput := errorDefaultMsg + `
	at [13:12]: <toplevel>
	at [0:0]: add_service`
	require.Equal(t, expectedOutput, errorToSerialize.Error())
}

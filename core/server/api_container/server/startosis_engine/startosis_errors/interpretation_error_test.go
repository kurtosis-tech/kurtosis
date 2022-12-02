package startosis_errors

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	dummyFilenameForTesting = "dummy.star"
)

func TestInterpretationError_serializationSimpleError(t *testing.T) {
	errorString := "This is an error!"
	errorToSerialize := NewInterpretationError(errorString)

	require.Equal(t, errorString, errorToSerialize.Error())
}

func TestInterpretationError_serializationWithCustomMsg(t *testing.T) {
	errorToSerialize := NewInterpretationErrorWithCustomMsg(
		[]CallFrame{
			*NewCallFrame("<toplevel>", NewScriptPosition(dummyFilenameForTesting, 13, 12)),
			*NewCallFrame("add_datastore_service", NewScriptPosition(dummyFilenameForTesting, 18, 16)),
			*NewCallFrame("add_service", NewScriptPosition(dummyFilenameForTesting, 0, 0)),
		},
		"Evaluation error: Missing `image` as part of the struct object",
	)

	expectedOutput := `Evaluation error: Missing ` + "`image`" + ` as part of the struct object
	at [` + dummyFilenameForTesting + `:13:12]: <toplevel>
	at [` + dummyFilenameForTesting + `:18:16]: add_datastore_service
	at [` + dummyFilenameForTesting + `:0:0]: add_service`
	require.Equal(t, expectedOutput, errorToSerialize.Error())
}

func TestInterpretationError_serializationFromStacktrace(t *testing.T) {
	errorToSerialize := NewInterpretationErrorFromStacktrace(
		[]CallFrame{
			*NewCallFrame("<toplevel>", NewScriptPosition(dummyFilenameForTesting, 13, 12)),
			*NewCallFrame("add_service", NewScriptPosition(dummyFilenameForTesting, 0, 0)),
		},
	)

	expectedOutput := errorDefaultMsg + `
	at [` + dummyFilenameForTesting + `:13:12]: <toplevel>
	at [` + dummyFilenameForTesting + `:0:0]: add_service`
	require.Equal(t, expectedOutput, errorToSerialize.Error())
}

func TestInterpretationError_WithCausedBy(t *testing.T) {
	rootCause := errors.New("root cause error")
	levelOneInterpretationError := WrapWithInterpretationError(rootCause, "This is the root interpretation error")
	userVisibleInterpretationError := WrapWithInterpretationError(levelOneInterpretationError, "An error happened!")

	expectedErrorMessage := `An error happened!
	Caused by: This is the root interpretation error
	Caused by: root cause error`
	require.Equal(t, expectedErrorMessage, userVisibleInterpretationError.Error())
}

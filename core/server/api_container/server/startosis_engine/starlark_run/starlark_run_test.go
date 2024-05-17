package starlark_run

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStarlarkRunMarshallers(t *testing.T) {
	packageId := "package-id-test"
	serializedScript := `
def run(plan, args):
	get_recipe = GetHttpRequestRecipe(
		port_id = args["port_id"],
		endpoint = args["endpoint"],
	)
	plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval=args["interval"], timeout=args["timeout"], service_name=args["service_name"])
`
	serializedParams := `{"greetings": "bonjour!"}`
	parallelism := int32(2)
	relativePathToMainFile := ""
	mainFunctionName := "main"
	experimentalFeatures := []int32{2, 7}
	restartPolicy := int32(1)
	initialSerializedParams := `{"bonjour": "foo"}`

	originalStarlarkRun := NewStarlarkRun(
		packageId,
		serializedScript,
		serializedParams,
		parallelism,
		relativePathToMainFile,
		mainFunctionName,
		experimentalFeatures,
		restartPolicy,
		initialSerializedParams,
	)

	marshaledStarlarkRun, err := json.Marshal(originalStarlarkRun)
	require.NoError(t, err)
	require.NotNil(t, marshaledStarlarkRun)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newStarlarkRun := &StarlarkRun{}

	err = json.Unmarshal(marshaledStarlarkRun, newStarlarkRun)
	require.NoError(t, err)

	require.EqualValues(t, originalStarlarkRun, newStarlarkRun)
}

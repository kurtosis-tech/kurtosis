package recipe

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecRecipe_String(t *testing.T) {
	serviceId := service.ServiceID("serviceId")
	commands := []string{"echo", "run"}

	expectedRecipeString := `ExecRecipe(service_id="serviceId", command="[\"echo\", \"run\"]")`
	execRecipe := NewExecRecipe(serviceId, commands)
	execRecipeString := execRecipe.String()
	require.Equal(t, expectedRecipeString, execRecipeString)
}

package recipe

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecRecipe_String(t *testing.T) {
	serviceName := service.ServiceName("serviceName")
	commands := []string{"echo", "run"}

	expectedRecipeString := `ExecRecipe(service_name="serviceName", command=["echo", "run"])`
	execRecipe := NewExecRecipe(serviceName, commands)
	execRecipeString := execRecipe.String()
	require.Equal(t, expectedRecipeString, execRecipeString)
}

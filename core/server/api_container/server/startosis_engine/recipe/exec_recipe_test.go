package recipe

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecRecipe_String(t *testing.T) {
	commands := []string{"echo", "run"}

	expectedRecipeString := `ExecRecipe(command=["echo", "run"])`
	execRecipe := NewExecRecipe(commands)
	execRecipeString := execRecipe.String()
	require.Equal(t, expectedRecipeString, execRecipeString)
}

package service_config

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"go.starlark.net/starlark"
)

//ReadyConditions holds all the information needed for ensuring service readiness
type ReadyConditions struct {
	recipe      recipe.Recipe
	field       string
	assertion   string
	targetValue starlark.Comparable
	interval    string
	timeout     string
}

package recipe_executor

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
)

type RecipeExecutor struct {
	recipeMap         map[string]*HttpRequestRecipe
	recipeToResultMap map[string]string
	recipeResultMap   map[string]*HttpRequestRuntimeValue
}

func NewRecipeExecutor() *RecipeExecutor {
	return &RecipeExecutor{
		recipeMap:         make(map[string]*HttpRequestRecipe),
		recipeResultMap:   make(map[string]*HttpRequestRuntimeValue),
		recipeToResultMap: make(map[string]string),
	}
}

func (re *RecipeExecutor) SaveRecipe(recipe *HttpRequestRecipe) string {
	uuid, _ := uuid_generator.GenerateUUIDString()
	re.recipeMap[uuid] = recipe
	return uuid
}

func (re *RecipeExecutor) CreateValue(uuid string) string {
	resultUuid, _ := uuid_generator.GenerateUUIDString()
	re.recipeToResultMap[resultUuid] = uuid
	return resultUuid
}

func (re *RecipeExecutor) GetRecipe(uuid string) *HttpRequestRecipe {
	return re.recipeMap[uuid]
}

func (re *RecipeExecutor) ExecuteAndSaveValue(ctx context.Context, serviceNetwork service_network.ServiceNetwork, uuid string) error {
	value, err := re.recipeMap[uuid].Execute(ctx, serviceNetwork)
	if err == nil {
		re.recipeResultMap[uuid] = value
	}
	return err
}

func (re *RecipeExecutor) GetValue(uuid string) *HttpRequestRuntimeValue {
	return re.recipeResultMap[uuid]
}

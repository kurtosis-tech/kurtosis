package recipe_executor

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
)

type RecipeExecutor struct {
	recipeMap       map[string]*HttpRequestRecipe
	recipeResultMap map[string]map[string]string
}

func NewRecipeExecutor() *RecipeExecutor {
	return &RecipeExecutor{
		recipeMap:       make(map[string]*HttpRequestRecipe),
		recipeResultMap: make(map[string]map[string]string),
	}
}

func (re *RecipeExecutor) CreateValue(recipe *HttpRequestRecipe) string {
	uuid, _ := uuid_generator.GenerateUUIDString()
	re.recipeMap[uuid] = recipe
	re.recipeResultMap[uuid] = nil
	return uuid
}

func (re *RecipeExecutor) GetRecipe(uuid string) *HttpRequestRecipe {
	return re.recipeMap[uuid]
}

func (re *RecipeExecutor) ExecuteValue(ctx context.Context, serviceNetwork service_network.ServiceNetwork, uuid string) error {
	value, err := re.recipeMap[uuid].Execute(ctx, serviceNetwork)
	if err != nil {
		return err
	}
	re.recipeResultMap[uuid] = value
	return nil
}

func (re *RecipeExecutor) GetValue(uuid string) map[string]string {
	return re.recipeResultMap[uuid]
}

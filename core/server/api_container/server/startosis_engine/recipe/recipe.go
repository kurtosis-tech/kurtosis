package recipe

import (
	"context"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

type Recipe interface {
	Execute(
		ctx context.Context,
		serviceNetwork service_network.ServiceNetwork,
		store *runtime_value_store.RuntimeValueStore,
		service *service.Service,
	) (map[string]starlark.Comparable, error)
	CreateStarlarkReturnValue(resultUuid string) (*starlark.Dict, *startosis_errors.InterpretationError)
	ResultMapToString(resultMap map[string]starlark.Comparable) string
	GetStarlarkReturnValuesAsStringList(resultUuid string) ([]string, *startosis_errors.InterpretationError)
}

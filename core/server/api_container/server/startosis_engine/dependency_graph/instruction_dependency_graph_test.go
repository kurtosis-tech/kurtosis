package dependency_graph

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
)

func TestAddDependency(t *testing.T) {
	_ = NewInstructionDependencyGraph([]types.ScheduledInstructionUuid{})
}

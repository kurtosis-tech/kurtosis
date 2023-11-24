package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
)

const ()

type PortForwardManager struct {
	kurtosis *kurtosis_context.KurtosisContext
}

func NewPortForwardManager(kurtosisContext *kurtosis_context.KurtosisContext) *PortForwardManager {
	return &PortForwardManager{
		kurtosis: kurtosisContext,
	}
}

// TODO(omar): get enclaves can take a moment so look for a lighter ping that also verifies we've an engine connection
// or consider an alternative health indicator
func (manager *PortForwardManager) Ping(ctx context.Context) error {
	_, err := manager.kurtosis.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Port Forward Manager failed to contact Kurtosis Engine")
	}
	return nil
}

func (manager *PortForwardManager) ForwardUserServiceToEphemeralPort(ctx context.Context, enclaveId string, serviceId string, portId string) (uint16, error) {
	return 0, nil
}

func (manager *PortForwardManager) ForwardUserServiceToStaticPort(ctx context.Context, enclaveId string, serviceId string, portId string, localPortNumber uint16) (uint16, error) {
	return 0, nil
}

package port_forward_manager

import "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"

const ()

type PortForwardManager struct {
	kurtosisContext *kurtosis_context.KurtosisContext
}

func NewPortForwardManager(kurtosisContext *kurtosis_context.KurtosisContext) *PortForwardManager {
	return &PortForwardManager{
		kurtosisContext: kurtosisContext,
	}
}

func (manager *PortForwardManager) Ping() error {
	// TODO(omar): check engine
	return nil
}

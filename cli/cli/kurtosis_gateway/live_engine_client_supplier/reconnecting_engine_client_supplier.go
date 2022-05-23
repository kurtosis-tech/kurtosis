package live_engine_client_supplier

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"time"
)

const (
	pollInterval = 2 * time.Second
)

type currentEngineInfo struct {
	EngineID

	proxyConn *connection.GatewayConnectionToKurtosis

	grpcConn net.Conn

	// This _might_ stateful, which is why we keep it here and hand
	// it back to requesters rather than creating a new one each time
	engineClient *kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// Class that will constantly poll the Kubernetes backend to try and find a live engine
type LiveEngineClientSupplier struct {
	kubernetesBackend *kubernetes.KubernetesKurtosisBackend

	currentInfo *currentEngineInfo

	stopUpdaterSignalChan chan interface{}

}

func NewLiveEngineClientSupplier(
	kubernetesKurtosisBackend *kubernetes.KubernetesKurtosisBackend,
	connectionProvider *connection.GatewayConnectionProvider,
) *LiveEngineClientSupplier {
	return &LiveEngineClientSupplier{
		kubernetesBackend:     kubernetesKurtosisBackend,
		currentInfo:           nil,
		stopUpdaterSignalChan: nil,
	}
}

func (supplier *LiveEngineClientSupplier) Start() error {
	if supplier.stopUpdaterSignalChan != nil {
		return stacktrace.NewError("Cannot start live engine client supplier because it's already started")
	}

	stopUpdaterSignalChan := make(chan interface{})

	go func() {
		poller := time.NewTicker(pollInterval)
		defer poller.Stop()

		for {
			select {
			case <- poller:




			case <-stopUpdaterSignalChan:
				return
			}
		}
	}()
	shouldStopUpdater := true
	defer func() {
		if shouldStopUpdater {
			stopUpdaterSignalChan <- nil
		}
	}()

	supplier.stopUpdaterSignalChan = stopUpdaterSignalChan

	shouldStopUpdater = false
	return nil
}

// NOTE: Do not save this value!! Just use it as a point-in-time piece of info
func GetEngineClient() (*kurtosis_engine_rpc_api_bindings.EngineServiceClient, error) {

}

func Close() {

}

func (supplier *kubernetes.KubernetesKurtosisBackend) checkIfWeShouldReplaceEngine() {


}
package live_engine_client_supplier

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	pollInterval = 2 * time.Second
)

type engineInfo struct {
	engineGuid engine.EngineGUID

	proxyConn connection.GatewayConnectionToKurtosis

	grpcConn net.Conn

	// This _might_ stateful, which is why we keep it here and hand
	// it back to requesters rather than creating a new one each time
	engineClient *kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// Class that will constantly poll the Kubernetes backend to try and find a live engine
type LiveEngineClientSupplier struct {
	kubernetesBackend *kubernetes.KubernetesKurtosisBackend

	connectionProvider *connection.GatewayConnectionProvider

	currentInfo *engineInfo

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
			case <- poller.C:
				supplier.replaceEngineIfNecessary()
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

func (supplier *LiveEngineClientSupplier) Stop() {
	if supplier.stopUpdaterSignalChan == nil {
		return
	}


}

// This function will gather the current state of engines in the cluster and compare it with the current state of the
// supplier. If necessary, the supplier's currently-tracked engine will be supplanted with the new running engine.
// NOT thread-safe!
func (supplier *LiveEngineClientSupplier) replaceEngineIfNecessary() {
	runningEngineFilters := &engine.EngineFilters{
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Running: true,
		},
	}
	runningEngines, err := supplier.kubernetesBackend.GetEngines(context.Background(), runningEngineFilters)
	if err != nil {
		logrus.Errorf("An error occurred finding running engines:\n%v", err)
		return
	}
	if len(runningEngines) == 0 {
		// If no running engines, don't give out EngineClients (because they won't work)
		closeEngineInfo(supplier.currentInfo)
		supplier.currentInfo = nil
		return
	}
	if len(runningEngines) > 1 {
		logrus.Errorf("Found > 1 engine running, which should never happen")
		return
	}

	var runningEngine *engine.Engine
	for _, runningEngine = range runningEngines {}

	// If we have no engine, we'll take anything we can get
	if supplier.currentInfo == nil {
		supplier.replaceCurrentEngineInfoBestEffort(runningEngine)
		return
	}

	// No need to replace if it's the one we're already connected to
	if supplier.currentInfo.engineGuid == runningEngine.GetGUID() {
		return
	}

	// If we get here, we must: a) have an engine that b) doesn't match the currently-running engine
	// Therefore, replace
	supplier.replaceCurrentEngineInfoBestEffort(runningEngine)
}

func (supplier *LiveEngineClientSupplier) replaceCurrentEngineInfoBestEffort(newEngine *engine.Engine) {
	newProxyConn := supplier.connectionProvider.ForEngine(newEngine)

	currentInfo := supplier.currentInfo

	currentProxy := currentInfo.proxyConn
	currentGrpc := currentInfo.grpcConn

	shouldCloseCurrentResources := false
	defer func() {
		if shouldCloseCurrentResources {
			// Very important not to pass in supplier.currentInfo, else we'll close the new info after replacement!
			closeEngineInfo(currentInfo)
		}
	}()

	shouldDestroyNewProxy := true
	defer func() {
		if shouldDestroyNewProxy {
			newProxyConn.Stop()
		}
	}()


	newEngineInfo := &engineInfo{
		engineGuid:   currentInfo.engineGuid,
		proxyConn:    nil,
		grpcConn:     nil,
		engineClient: nil,
	}
	supplier.currentInfo = newEngineInfo

	shouldDestroyNewProxy = false
}

func closeEngineInfo(info *engineInfo) {
	// Ordering is important
	info.grpcConn.Close()
	info.proxyConn.Stop()
}
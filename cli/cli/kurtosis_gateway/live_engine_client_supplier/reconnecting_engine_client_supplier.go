package live_engine_client_supplier

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

const (
	pollInterval = 2 * time.Second
)

type engineInfo struct {
	engineGuid engine.EngineGUID

	proxyConn connection.GatewayConnectionToKurtosis

	grpcConn *grpc.ClientConn

	// This _might_ stateful, which is why we keep it here and hand
	// it back to requesters rather than creating a new one each time
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// Class that will constantly poll the Kubernetes backend to try and find a live engine
type LiveEngineClientSupplier struct {
	kubernetesBackend backend_interface.KurtosisBackend

	connectionProvider *connection.GatewayConnectionProvider

	currentInfo *engineInfo

	stopUpdaterSignalChan chan interface{}
}

func NewLiveEngineClientSupplier(
	kurtosisBackend backend_interface.KurtosisBackend,
	connectionProvider *connection.GatewayConnectionProvider,
) *LiveEngineClientSupplier {
	return &LiveEngineClientSupplier{
		kubernetesBackend:     kurtosisBackend,
		currentInfo:           nil,
		stopUpdaterSignalChan: nil,
		connectionProvider:    connectionProvider,
	}
}

func (supplier *LiveEngineClientSupplier) Start() error {
	if supplier.stopUpdaterSignalChan != nil {
		return stacktrace.NewError("Cannot start live engine client supplier because it's already started")
	}
	// Get connected to the engine
	supplier.replaceEngineIfNecessary()

	// Start health checking the engine
	stopUpdaterSignalChan := make(chan interface{})
	go func() {
		poller := time.NewTicker(pollInterval)
		defer poller.Stop()

		for {
			select {
			case <-poller.C:
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
func (supplier *LiveEngineClientSupplier) GetEngineClient() (kurtosis_engine_rpc_api_bindings.EngineServiceClient, error) {
	if supplier.currentInfo == nil {
		return nil, stacktrace.NewError("Expected to have info about a running live engine, instead no info was found")
	}
	return supplier.currentInfo.engineClient, nil
}

func (supplier *LiveEngineClientSupplier) Stop() {
	if supplier.stopUpdaterSignalChan == nil {
		return
	}
	supplier.stopUpdaterSignalChan <- nil
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
	for _, onlyEngineInMap := range runningEngines {
		runningEngine = onlyEngineInMap
	}

	// If we have no engine client, we'll take anything we can get
	if supplier.currentInfo == nil {
		if err := supplier.replaceCurrentEngineInfo(runningEngine); err != nil {
			logrus.Errorf("Expected to be able to replace the current engine info, instead a non-nil error was returned:\n%v", err)
		}
		return
	}

	// No need to replace if it's the one we're already connected to
	if supplier.currentInfo.engineGuid == runningEngine.GetGUID() {
		return
	}

	// If we get here, we must: a) have an engine that b) doesn't match the currently-running engine
	// Therefore, replace
	if err := supplier.replaceCurrentEngineInfo(runningEngine); err != nil {
		logrus.Errorf("Expected to be able to replace the current engine info, instead a non-nil error was returned:\n%v", err)
	}
}

func (supplier *LiveEngineClientSupplier) replaceCurrentEngineInfo(newEngine *engine.Engine) error {
	// Get a port forwarded connection to the new engine
	newProxyConn, err := supplier.connectionProvider.ForEngine(newEngine)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get a port forwarded connection to engine '%v', instead a non-nil error was returned", newEngine.GetGUID())
	}
	shouldDestroyNewProxy := true
	defer func() {
		if shouldDestroyNewProxy {
			newProxyConn.Stop()
		}
	}()
	// Create an engine client that sends requests to our new client
	newGrpcConnection, err := newProxyConn.GetGrpcClientConn()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get a GRPC client connection to engine '%v', instead a non-nil error was returned", newEngine.GetGUID())
	}
	shouldDestroyNewGrpcClientConn := true
	defer func() {
		if shouldDestroyNewGrpcClientConn {
			newGrpcConnection.Close()
		}
	}()
	newEngineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(newGrpcConnection)
	newEngineInfo := &engineInfo{
		engineGuid:   newEngine.GetGUID(),
		proxyConn:    newProxyConn,
		grpcConn:     newGrpcConnection,
		engineClient: newEngineClient,
	}

	oldEngineInfo := supplier.currentInfo
	shouldReinsertOldEngineInfo := true
	defer func() {
		// Put back oldEngineInfo if a failure occurs
		if shouldReinsertOldEngineInfo {
			supplier.currentInfo = oldEngineInfo
		} else {
			// If we're not going to reinsert the old engine info, then close it as it's no longer needed
			closeEngineInfo(oldEngineInfo)
		}
	}()
	supplier.currentInfo = newEngineInfo
	logrus.Infof("Connected to running Kurtosis engine with id '%v'", newEngine.GetGUID())
	shouldDestroyNewProxy = false
	shouldDestroyNewGrpcClientConn = false
	shouldReinsertOldEngineInfo = false
	return nil
}

func closeEngineInfo(info *engineInfo) {
	if info == nil {
		return
	}
	// Ordering is important
	info.grpcConn.Close()
	info.proxyConn.Stop()
}

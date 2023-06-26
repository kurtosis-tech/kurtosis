package enclave_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	oneEnclave = 1

	// This is the same default value used in the `kurtosis enclave add` CLI command
	defaultApiContainerLogLevel = logrus.DebugLevel

	// The partitioning feature is disabled by default because the Enclave Pool
	// is enabled only for K8s and the network partitioning feature is not implemented yet
	defaultIsPartitioningEnabled = false

	// I think this value is deprecated on the CreateEnclave method //TODO I should check it
	defaultMetricsUserID = "nothing"
	// I think this value is deprecated on the CreateEnclave method //TODO I should check it
	defaultDidUserAcceptSendingMetrics = false

	maxRetryAfterFails = 60

	timeBetweenRetries = time.Second * 1
)

type EnclavePool struct {
	idleEnclavesChan chan *kurtosis_engine_rpc_api_bindings.EnclaveInfo
	enclaveCreator   *EnclaveCreator
	engineVersion    string
}

func CreateEnclavePool(
	enclaveCreator *EnclaveCreator,
	poolSize uint8,
	engineVersion string,
) *EnclavePool {

	// The amount of idle enclaves is equal to the chan capacity + one enclave
	// that will be waiting until the channel is unblocked
	chanCapacity := poolSize - oneEnclave

	idleEnclavesChan := make(chan *kurtosis_engine_rpc_api_bindings.EnclaveInfo, chanCapacity)

	enclavePool := &EnclavePool{
		idleEnclavesChan: idleEnclavesChan,
		enclaveCreator:   enclaveCreator,
		engineVersion:    engineVersion,
	}

	go enclavePool.run()

	return enclavePool
}

// TODO we have to put the K8s condition in some place
// TODO we have to evaluate if users is requesting an enclave with defaults value or not
func (pool *EnclavePool) GetEnclave(newEnclaveName string) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	// TODO Rename enclave
	// TODO change the logLevel value
	// TODO what we do with the Metrics ID value
	// TODO we should also check if the user is requesting for partitioning enable
	// TODO we should check the enclave status before returning it

	// TODO el problema mas importante que encuentro ahora es que no hay comunicación entre el
	// TODO EngineServer y el APIContainerServer como para actualizar el APIContainer logLevel value
	// TODO desde el Engine

	enclaveInfo, ok := <-pool.idleEnclavesChan
	if !ok {
		//TODO improve the error here
		return nil, stacktrace.NewError("A new enclave can't be returned because the enclave pool channel is closed")
	}

	logrus.Debugf("Returning enclave Info '%+v' for requested enclave name '%s'", enclaveInfo, newEnclaveName)

	return enclaveInfo, nil
}

// The enclave pool feature is only available for Kubernetes so far and it will be activated
// only if users require this when setting the pool-size value
func isEnclavePoolAllowedForThisConfig(
	poolSize uint8,
	kurtosisBackendType args.KurtosisBackendType,
) bool {

	if poolSize > 0 && kurtosisBackendType == args.KurtosisBackendType_Kubernetes {
		return true
	}

	return false
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================

// Run will be executed in a sub-routine and will be in charge of:
// 1-
func (pool *EnclavePool) run() {
	//TODO first iterate on all the existing enclaves in order to find idle enclaves already created
	//TODO if there is any compare version and add it into the chan or discard them
	defer close(pool.idleEnclavesChan)

	fails := 0

	for {

		newEnclaveInfo, err := pool.createNewEnclave()
		if err != nil {
			// Retry strategy if something fails
			logrus.Errorf("An error occurred creating a new idle enclave. Error\n%v", err)
			fails++
			if fails >= maxRetryAfterFails {
				//TODO log something
				break
			}
			time.Sleep(timeBetweenRetries)
			//TODO put some debug log
		}

		pool.idleEnclavesChan <- newEnclaveInfo
		logrus.Debugf("Enclave '%+v' was added intho the pool channel", newEnclaveInfo)
	}
}

func (pool *EnclavePool) createNewEnclave() (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	ctx := context.Background()

	enclaveName, err := GetRandomIdleEnclaveName()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred generating a random name for a new idle enclave.",
		)
	}

	apiContainerVersion := pool.engineVersion

	newEnclaveInfo, err := pool.enclaveCreator.CreateEnclave(
		ctx,
		apiContainerVersion,
		defaultApiContainerLogLevel,
		enclaveName,
		defaultIsPartitioningEnabled,
	)
	if err != nil {
		//TODO complete the message adding more values
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating idle enclave with name '%s'",
			enclaveName,
		)
	}

	logrus.Debugf("New idle enclave created '%+v'", newEnclaveInfo)
	return newEnclaveInfo, nil
}

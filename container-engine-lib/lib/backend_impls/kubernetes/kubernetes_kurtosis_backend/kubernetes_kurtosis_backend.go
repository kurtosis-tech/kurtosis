package kubernetes_kurtosis_backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
)

type KubernetesKurtosisBackend struct {
	kubernetesManager *kubernetes_manager.KubernetesManager

	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider

	cliModeArgs *shared_helpers.CliModeArgs

	engineServerModeArgs *shared_helpers.EngineServerModeArgs

	// Will only be filled out for the API container
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs
}

func (backend *KubernetesKurtosisBackend) GetEngineLogs(ctx context.Context, outputDirpath string) error {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DumpKurtosis(ctx context.Context, outputDirpath string) error {
	//TODO implement me
	panic("implement me")
}

// Private constructor that the other public constructors will use
func newKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	cliModeArgs *shared_helpers.CliModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:    kubernetesManager,
		objAttrsProvider:     objAttrsProvider,
		cliModeArgs:          cliModeArgs,
		engineServerModeArgs: engineServerModeArgs,
		apiContainerModeArgs: apiContainerModeArgs,
	}
}

func NewAPIContainerKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	ownEnclaveUuid enclave.EnclaveUUID,
	ownNamespaceName string,
) *KubernetesKurtosisBackend {
	modeArgs := shared_helpers.NewApiContainerModeArgs(ownEnclaveUuid, ownNamespaceName)
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		nil,
		nil,
		modeArgs,
	)
}

func NewEngineServerKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
) *KubernetesKurtosisBackend {
	modeArgs := &shared_helpers.EngineServerModeArgs{}
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		nil,
		modeArgs,
		nil,
	)
}

func NewCLIModeKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
) *KubernetesKurtosisBackend {
	modeArgs := &shared_helpers.CliModeArgs{}
	return newKubernetesKurtosisBackend(
		kubernetesManager,
		modeArgs,
		nil,
		nil,
	)
}

func NewKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	// TODO Remove the necessity for these different args by splitting the *KubernetesKurtosisBackend into multiple
	//  backends per consumer, e.g. APIContainerKurtosisBackend, CLIKurtosisBackend, EngineKurtosisBackend, etc.
	//  This can only happen once the CLI no longer uses the same functionality as API container, engine, etc. though
	cliModeArgs *shared_helpers.CliModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	apiContainerModeargs *shared_helpers.ApiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:    kubernetesManager,
		objAttrsProvider:     objAttrsProvider,
		cliModeArgs:          cliModeArgs,
		engineServerModeArgs: engineServerModeArgs,
		apiContainerModeArgs: apiContainerModeargs,
	}
}

func (backend *KubernetesKurtosisBackend) FetchImage(ctx context.Context, image string) error {
	logrus.Warnf("FetchImage isn't implemented for Kubernetes yet")
	return nil
}

func (backend *KubernetesKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {
	kubernetesEngine, err := engine_functions.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		envVars,
		backend.kubernetesManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating engine using image '%v:%v', grpc port number '%v' and environment variables '%+v'",
			imageOrgAndRepo,
			imageVersionTag,
			grpcPortNum,
			envVars,
		)
	}
	return kubernetesEngine, nil
}

func (backend *KubernetesKurtosisBackend) GetEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (map[engine.EngineGUID]*engine.Engine, error) {
	engines, err := engine_functions.GetEngines(ctx, filters, backend.kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters '%+v'", filters)
	}
	return engines, nil
}

func (backend *KubernetesKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	successfulEngineGuids, erroredEngineGuids, err := engine_functions.StopEngines(ctx, filters, backend.kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping engines using filters '%+v'", filters)
	}
	return successfulEngineGuids, erroredEngineGuids, nil
}

func (backend *KubernetesKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	successfulEngineGuids, erroredEngineGuids, err := engine_functions.DestroyEngines(ctx, filters, backend.kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters '%+v'", filters)
	}
	return successfulEngineGuids, erroredEngineGuids, nil
}

func (backend *KubernetesKurtosisBackend) RegisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceName]bool) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	successfullyRegisteredService, failedServices, err := user_services_functions.RegisterUserServices(
		ctx,
		enclaveUuid,
		services,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
	if err != nil {
		var serviceIds []service.ServiceName
		for serviceId := range services {
			serviceIds = append(serviceIds, serviceId)
		}
		return nil, nil, stacktrace.Propagate(err, "Unexpected error registering services with Names '%v' to enclave '%s'", serviceIds, enclaveUuid)
	}
	return successfullyRegisteredService, failedServices, nil
}

func (backend *KubernetesKurtosisBackend) UnregisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	successfullyUnregisteredServices, failedServices, err := user_services_functions.UnregisterUserServices(
		ctx,
		enclaveUuid,
		services,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
	if err != nil {
		var serviceUuids []service.ServiceUUID
		for serviceUuid := range services {
			serviceUuids = append(serviceUuids, serviceUuid)
		}
		return nil, nil, stacktrace.Propagate(err, "Unexpected error unregistering services with GUIDs '%v' from enclave '%s'", serviceUuids, enclaveUuid)
	}
	return successfullyUnregisteredServices, failedServices, nil
}

func (backend *KubernetesKurtosisBackend) StartRegisteredUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]*service.ServiceConfig,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	successfullyStartedServices, failedServices, err := user_services_functions.StartRegisteredUserServices(
		ctx,
		enclaveUuid,
		services,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
	if err != nil {
		var serviceUuids []service.ServiceUUID
		for serviceUuid := range services {
			serviceUuids = append(serviceUuids, serviceUuid)
		}
		return nil, nil, stacktrace.Propagate(err, "Unexpected error starting services with GUIDs '%v' in enclave '%s'", serviceUuids, enclaveUuid)
	}
	return successfullyStartedServices, failedServices, nil
}

func (backend *KubernetesKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (successfulUserServices map[service.ServiceUUID]*service.Service, resultError error) {
	return user_services_functions.GetUserServices(
		ctx,
		enclaveUuid,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (successfulUserServiceLogs map[service.ServiceUUID]io.ReadCloser, erroredUserServiceUuids map[service.ServiceUUID]error, resultError error) {
	return user_services_functions.GetUserServiceLogs(
		ctx,
		enclaveUuid,
		filters,
		shouldFollowLogs,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceId service.ServiceUUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because pausing is not supported by Kubernetes", serviceId, enclaveUuid)
}

func (backend *KubernetesKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceId service.ServiceUUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because unpausing is not supported by Kubernetes", serviceId, enclaveUuid)
}

// TODO Switch these to streaming methods, so that huge command outputs don't blow up the memory of the API container
func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceUUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceUUID]*exec_result.ExecResult,
	erroredUserServiceUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	return user_services_functions.RunUserServiceExecCommands(
		ctx,
		enclaveUuid,
		userServiceCommands,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUUID service.ServiceUUID) (resultConn net.Conn, resultErr error) {
	// See https://github.com/kubernetes/client-go/issues/912
	/*
		in := streams.NewIn(os.Stdin)
		if err := in.SetRawTerminal(); err != nil{
					 // handle err
		}
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:             in,
			Stdout:           stdout,
			Stderr:            stderr,
		}
	*/

	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Getting a connection with a user service isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPath string,
	output io.Writer,
) error {
	return user_services_functions.CopyFilesFromUserService(
		ctx,
		enclaveUuid,
		serviceUuid,
		srcPath,
		output,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceUUID]bool, resultErroredGuids map[service.ServiceUUID]error, resultErr error) {
	return user_services_functions.StopUserServices(
		ctx,
		enclaveUuid,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceUUID]bool, resultErroredGuids map[service.ServiceUUID]error, resultErr error) {
	return user_services_functions.DestroyUserServices(
		ctx,
		enclaveUuid,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) CreateLogsDatabase(ctx context.Context, logsDatabaseHttpPortNumber uint16) (*logs_database.LogsDatabase, error) {
	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Creating the logs database isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) GetLogsDatabase(
	ctx context.Context,
) (*logs_database.LogsDatabase, error) {
	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Getting the logs database isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) DestroyLogsDatabase(
	ctx context.Context,
) error {
	// TODO IMPLEMENT
	return stacktrace.NewError("Destroying the logs database isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) CreateLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, logsCollectorHttpPortNumber uint16, logsCollectorTcpPortNumber uint16) (*logs_collector.LogsCollector, error) {
	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Creating the logs collector isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) GetLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*logs_collector.LogsCollector, error) {
	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Creating the logs collector isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) DestroyDeprecatedCentralizedLogsResources(ctx context.Context) error {
	logrus.Debugf("Destroy the deprecated centralized logs resources is not needed for Kubernetes")
	return nil
}

func (backend *KubernetesKurtosisBackend) DestroyLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) error {
	// TODO IMPLEMENT
	return stacktrace.NewError("Destroy the logs collector for enclave isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) DestroyLogsCollector(ctx context.Context) error {
	// TODO IMPLEMENT
	return stacktrace.NewError("Destroying the logs collector isn't yet implemented on Kubernetes")
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getEnclaveNamespaceName(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (string, error) {
	// TODO This is a big janky hack that results from *KubernetesKurtosisBackend containing functions for all of API containers, engines, and CLIs
	//  We want to fix this by splitting the *KubernetesKurtosisBackend into a bunch of different backends, one per user, but we can only
	//  do this once the CLI no longer uses API container functionality (e.g. GetServices)
	// CLIs and engines can list namespaces so they'll be able to use the regular list-namespaces-and-find-the-one-matching-the-enclave-ID
	// API containers can't list all namespaces due to being namespaced objects themselves (can only view their own namespace, so
	// they can only check if the requested enclave matches the one they have stored
	var namespaceName string
	if backend.cliModeArgs != nil || backend.engineServerModeArgs != nil {
		matchLabels := getEnclaveMatchLabels()
		matchLabels[label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString()] = string(enclaveUuid)

		namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
		}

		numOfNamespaces := len(namespaces.Items)
		if numOfNamespaces == 0 {
			return "", stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
		}
		if numOfNamespaces > 1 {
			return "", stacktrace.NewError("Expected to find only one enclave namespace matching enclave ID '%v', but found '%v'; this is a bug in Kurtosis", enclaveUuid, numOfNamespaces)
		}

		namespaceName = namespaces.Items[0].Name
	} else if backend.apiContainerModeArgs != nil {
		if enclaveUuid != backend.apiContainerModeArgs.GetOwnEnclaveId() {
			return "", stacktrace.NewError(
				"Received a request to get namespace for enclave '%v', but the Kubernetes Kurtosis backend is running in an API "+
					"container in a different enclave '%v' (so Kubernetes would throw a permission error)",
				enclaveUuid,
				backend.apiContainerModeArgs.GetOwnEnclaveId(),
			)
		}
		namespaceName = backend.apiContainerModeArgs.GetOwnNamespaceName()
	} else {
		return "", stacktrace.NewError("Received a request to get an enclave namespace's name, but the Kubernetes Kurtosis backend isn't in any recognized mode; this is a bug in Kurtosis")
	}

	return namespaceName, nil
}

func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
}

// This is a helper function that will take multiple errors, each identified by an ID, and format them together
// If no errors are returned, this function returns nil
func buildCombinedError(errorsById map[string]error, titleStr string) error {
	allErrorStrs := []string{}
	for errorId, stopErr := range errorsById {
		errorFormatStr := ">>>>>>>>>>>>> %v %v <<<<<<<<<<<<<\n" +
			"%v\n" +
			">>>>>>>>>>>>> END %v %v <<<<<<<<<<<<<"
		errorStr := fmt.Sprintf(
			errorFormatStr,
			strings.ToUpper(titleStr),
			errorId,
			stopErr.Error(),
			strings.ToUpper(titleStr),
			errorId,
		)
		allErrorStrs = append(allErrorStrs, errorStr)
	}

	if len(allErrorStrs) > 0 {
		// NOTE: This is one of the VERY rare cases where we don't want to use stacktrace.Propagate, because
		// attaching stack information for this method (which simply combines errors) just isn't useful. The
		// expected behaviour is that the caller of this function will use stacktrace.Propagate
		return errors.New(strings.Join(
			allErrorStrs,
			"\n\n",
		))
	}

	return nil
}

package kubernetes_kurtosis_backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
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


// Private constructor that the other public constructors will use
func newKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	cliModeArgs *shared_helpers.CliModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:               kubernetesManager,
		objAttrsProvider:                objAttrsProvider,
		cliModeArgs:                     cliModeArgs,
		engineServerModeArgs:            engineServerModeArgs,
		apiContainerModeArgs:            apiContainerModeArgs,
	}
}

func NewAPIContainerKubernetesKurtosisBackend(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	ownEnclaveId enclave.EnclaveID,
	ownNamespaceName string,
) *KubernetesKurtosisBackend {
	modeArgs := shared_helpers.NewApiContainerModeArgs(ownEnclaveId, ownNamespaceName)
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
	// TODO Remove the necessity for these different args by splitting the *KubernetesKurtosisBackend into multiple backends per consumer, e.g.
	//  APIContainerKurtosisBackend, CLIKurtosisBackend, EngineKurtosisBackend, etc. This can only happen once the CLI
	//  no longer uses the same functionality as API container, engine, etc. though
	cliModeArgs *shared_helpers.CliModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	apiContainerModeargs *shared_helpers.ApiContainerModeArgs,
) *KubernetesKurtosisBackend {
	objAttrsProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	return &KubernetesKurtosisBackend{
		kubernetesManager:               kubernetesManager,
		objAttrsProvider:                objAttrsProvider,
		cliModeArgs:                     cliModeArgs,
		engineServerModeArgs:            engineServerModeArgs,
		apiContainerModeArgs:            apiContainerModeargs,
	}
}

func (backend *KubernetesKurtosisBackend) PullImage(image string) error {
	return stacktrace.NewError("PullImage isn't implemented for Kubernetes yet")
}

func (backend KubernetesKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {
	engine, err := engine_functions.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		grpcProxyPortNum,
		envVars,
		backend.kubernetesManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating engine using image '%v:%v', grpc port number '%v', grpc proxy port number '%v' and environment variables '%+v'",
			imageOrgAndRepo,
			imageVersionTag,
			grpcPortNum,
			grpcProxyPortNum,
			envVars,
		)
	}
	return engine, nil
}

func (backend KubernetesKurtosisBackend) GetEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (map[engine.EngineGUID]*engine.Engine, error) {
	engines, err := engine_functions.GetEngines(ctx, filters, backend.kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters '%+v'", filters)
	}
	return engines, nil
}

func (backend KubernetesKurtosisBackend) StopEngines(
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

func (backend KubernetesKurtosisBackend) DestroyEngines(
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

func (backend *KubernetesKurtosisBackend) StartUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	services map[service.ServiceGUID]*service.ServiceConfig,
) (
		map[service.ServiceGUID]*service.Service,
		map[service.ServiceGUID]error,
		error,
) {
	return user_services_functions.StartUserServices(
		ctx,
		enclaveId,
		services,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	return user_services_functions.GetUserServices(
		ctx,
		enclaveId,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
	return user_services_functions.GetUserServiceLogs(
		ctx,
		enclaveId,
		filters,
		shouldFollowLogs,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because pausing is not supported by Kubernetes", serviceId, enclaveId)
}

func (backend *KubernetesKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because unpausing is not supported by Kubernetes", serviceId, enclaveId)
}

// TODO Switch these to streaming methods, so that huge command outputs don't blow up the memory of the API container
func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	return user_services_functions.RunUserServiceExecCommands(
		ctx,
		enclaveId,
		userServiceCommands,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGUID service.ServiceGUID) (resultConn net.Conn, resultErr error) {
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
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPath string,
	output io.Writer,
) error {
	return user_services_functions.CopyFilesFromUserService(
		ctx,
		enclaveId,
		serviceGuid,
		srcPath,
		output,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceGUID]bool, resultErroredGuids map[service.ServiceGUID]error, resultErr error) {
	return user_services_functions.StopUserServices(
		ctx,
		enclaveId,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceGUID]bool, resultErroredGuids map[service.ServiceGUID]error, resultErr error) {
	return user_services_functions.DestroyUserServices(
		ctx,
		enclaveId,
		filters,
		backend.cliModeArgs,
		backend.apiContainerModeArgs,
		backend.engineServerModeArgs,
		backend.kubernetesManager)
}

func (backend *KubernetesKurtosisBackend) getEnclaveNamespaceName(ctx context.Context, enclaveId enclave.EnclaveID) (string, error) {
	// TODO This is a big janky hack that results from *KubernetesKurtosisBackend containing functions for all of API containers, engines, and CLIs
	//  We want to fix this by splitting the *KubernetesKurtosisBackend into a bunch of different backends, one per user, but we can only
	//  do this once the CLI no longer uses API container functionality (e.g. GetServices)
	// CLIs and engines can list namespaces so they'll be able to use the regular list-namespaces-and-find-the-one-matching-the-enclave-ID
	// API containers can't list all namespaces due to being namespaced objects themselves (can only view their own namespace, so
	// they can only check if the requested enclave matches the one they have stored
	var namespaceName string
	if backend.cliModeArgs != nil || backend.engineServerModeArgs != nil {
		matchLabels := getEnclaveMatchLabels()
		matchLabels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()] = string(enclaveId)

		namespaces, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
		}

		numOfNamespaces := len(namespaces.Items)
		if numOfNamespaces == 0 {
			return "", stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
		}
		if numOfNamespaces > 1 {
			return "", stacktrace.NewError("Expected to find only one enclave namespace matching enclave ID '%v', but found '%v'; this is a bug in Kurtosis", enclaveId, numOfNamespaces)
		}

		namespaceName = namespaces.Items[0].Name
	} else if backend.apiContainerModeArgs != nil {
		if enclaveId != backend.apiContainerModeArgs.GetOwnEnclaveId() {
			return "", stacktrace.NewError(
				"Received a request to get namespace for enclave '%v', but the Kubernetes Kurtosis backend is running in an API " +
					"container in a different enclave '%v' (so Kubernetes would throw a permission error)",
				enclaveId,
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


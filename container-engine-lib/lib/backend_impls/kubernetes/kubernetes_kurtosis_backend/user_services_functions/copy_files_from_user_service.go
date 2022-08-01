package user_services_functions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	apiv1 "k8s.io/api/core/v1"
)

const (
	tarSuccessExitCode = 0
)

func CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPath string,
	output io.Writer,
	cliModeArgs *shared_functions.CliModeArgs,
	apiContainerModeArgs *shared_functions.ApiContainerModeArgs,
	engineServerModeArgs *shared_functions.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	namespaceName, err := shared_functions.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	objectAndResources, err := shared_functions.GetSingleUserServiceObjectsAndResources(ctx, enclaveId, serviceGuid, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service object & Kubernetes resources for service '%v' in enclave '%v'", serviceGuid, enclaveId)
	}
	pod := objectAndResources.KubernetesResources.Pod
	if pod == nil {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because no pod exists for the service",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}
	if pod.Status.Phase != apiv1.PodRunning {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because the pod isn't running",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}

	commandToRun := fmt.Sprintf(
		`if command -v 'tar' > /dev/null; then tar cf - '%v'; else echo "Cannot copy files from path '%v' because the tar binary doesn't exist on the machine" >&2; exit 1; fi`,
		srcPath,
		srcPath,
	)
	shWrappedCommandToRun := []string{
		"sh",
		"-c",
		commandToRun,
	}

	// NOTE: If we hit problems with very large files and connections breaking before they do, 'kubectl cp' implements a retry
	// mechanism that we could draw inspiration from:
	// https://github.com/kubernetes/kubectl/blob/335090af6913fb1ebf4a1f9e2463c46248b3e68d/pkg/cmd/cp/cp.go#L345
	stdErrOutput := &bytes.Buffer{}
	exitCode, err := kubernetesManager.RunExecCommand(
		namespaceName,
		pod.Name,
		userServiceContainerName,
		shWrappedCommandToRun,
		output,
		stdErrOutput,
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running command '%v' on pod '%v' for service '%v' in namespace '%v'",
			commandToRun,
			pod.Name,
			serviceGuid,
			namespaceName,
		)
	}
	if exitCode != tarSuccessExitCode {
		return stacktrace.NewError(
			"Command '%v' exited with non-%v exit code %v and the following STDERR:\n%v",
			commandToRun,
			tarSuccessExitCode,
			exitCode,
			stdErrOutput.String(),
		)
	}

	return nil
}
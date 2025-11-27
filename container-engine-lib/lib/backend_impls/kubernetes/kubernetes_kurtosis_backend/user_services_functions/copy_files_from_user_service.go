package user_services_functions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"path/filepath"
)

const (
	tarSuccessExitCode                   = 0
	doNotIncludeParentDirInArchiveSymbol = "*"
	ignoreParentDirInArchiveSymbol       = "."
)

// tries to copy contents of dir and if not dir reverts to default copy
var commandString = `if command -v 'tar' > /dev/null; then cd '%v' && (tar cf - -C '%v' . || tar cf - '%v'); else echo "Cannot copy files from path '%v' because the tar binary doesn't exist on the machine" >&2; exit 1; fi`

func CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPath string,
	output io.Writer,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	objectAndResources, err := shared_helpers.GetSingleUserServiceObjectsAndResources(ctx, enclaveId, serviceUuid, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service object & Kubernetes resources for service '%v' in enclave '%v'", serviceUuid, enclaveId)
	}
	pod := objectAndResources.KubernetesResources.Pod
	if pod == nil {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because no pod exists for the service",
			srcPath,
			serviceUuid,
			enclaveId,
		)
	}
	if pod.Status.Phase != apiv1.PodRunning {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because the pod isn't running",
			srcPath,
			serviceUuid,
			enclaveId,
		)
	}

	// we remove trailing slash
	srcPath = filepath.Clean(srcPath)
	// we get the base dir | file
	srcPathBase := filepath.Base(srcPath)
	// we get the dir that holds base the dir | file
	srcPathDir := filepath.Dir(srcPath)

	var commandToRun string

	if srcPathBase == doNotIncludeParentDirInArchiveSymbol {
		srcPathBase = ignoreParentDirInArchiveSymbol
	}
	commandToRun = fmt.Sprintf(
		commandString,
		srcPathDir,
		srcPathBase,
		srcPathBase,
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
			serviceUuid,
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

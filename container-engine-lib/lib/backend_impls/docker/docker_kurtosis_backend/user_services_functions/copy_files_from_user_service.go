package user_service_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"path/filepath"
)

const (
	doNotIncludeParentDirInArchiveSymbol = "*"
	ignoreParentDirInArchiveSymbolFormat = "%v/."
)

// It returns io.ReadCloser which is a tar stream. It's up to the caller to close the reader.
func CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPathOnContainer string,
	output io.Writer,
	dockerManager *docker_manager.DockerManager,
) error {

	srcPath := srcPathOnContainer
	srcPathBase := filepath.Base(srcPathOnContainer)
	if srcPathBase == doNotIncludeParentDirInArchiveSymbol {
		srcPath = filepath.Dir(srcPathOnContainer)
		srcPath = fmt.Sprintf(ignoreParentDirInArchiveSymbolFormat, srcPath)
	}

	logrus.Debugf("Copying contents from the src path: %v and base %v", srcPath, srcPathBase)
	_, serviceDockerResources, err := getSingleUserServiceObjAndResourcesNoMutex(ctx, enclaveId, serviceUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service with UUID '%v' in enclave with ID '%v'", serviceUuid, enclaveId)
	}
	container := serviceDockerResources.ServiceContainer

	tarStreamReadCloser, err := dockerManager.CopyFromContainer(ctx, container.GetId(), srcPath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying content from sourcepath '%v' in container '%v' for user service '%v' in enclave '%v'",
			srcPathOnContainer,
			container.GetName(),
			serviceUuid,
			enclaveId,
		)
	}
	defer tarStreamReadCloser.Close()

	if numBytesCopied, err := io.Copy(output, tarStreamReadCloser); err != nil {
		return stacktrace.Propagate(
			err,
			"'%v' bytes copied before an error occurred copying the bytes of TAR'd up files at '%v' on service '%v' to the output",
			numBytesCopied,
			srcPathOnContainer,
			serviceUuid,
		)
	}

	return nil
}

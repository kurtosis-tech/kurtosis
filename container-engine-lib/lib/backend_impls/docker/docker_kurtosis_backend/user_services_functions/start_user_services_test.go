package user_service_functions

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/stretchr/testify/require"
)

const testServiceUuid = "test-service-uuid"

var (
	errAddressInUse  = errors.New("failed to bind host port for 0.0.0.0:46006:172.16.0.28:9090/tcp: address already in use")
	errPortAllocated = errors.New("Bind for 0.0.0.0:46006 failed: port is already allocated")
	errUnrelated     = errors.New("no such image: some-image")
)

// fakeCreateAndStart returns the queued errors in order, then succeeds.
func fakeCreateAndStart(errs ...error) (createAndStartContainerFunc, *int) {
	callCount := 0
	countPtr := &callCount
	return func(_ context.Context, _ *docker_manager.CreateAndStartContainerArgs) (string, map[nat.Port]*nat.PortBinding, error) {
		callIdx := *countPtr
		*countPtr++
		if callIdx < len(errs) && errs[callIdx] != nil {
			return "", nil, errs[callIdx]
		}
		return "container-id", map[nat.Port]*nat.PortBinding{}, nil
	}, countPtr
}

func TestRetrySucceedsFirstTry(t *testing.T) {
	createAndStart, calls := fakeCreateAndStart()
	containerId, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, true, testServiceUuid)
	require.NoError(t, err)
	require.Equal(t, "container-id", containerId)
	require.Equal(t, 1, *calls)
}

func TestRetryRecoversFromAddressInUse(t *testing.T) {
	createAndStart, calls := fakeCreateAndStart(errAddressInUse, errAddressInUse)
	containerId, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, true, testServiceUuid)
	require.NoError(t, err)
	require.Equal(t, "container-id", containerId)
	require.Equal(t, 3, *calls)
}

func TestRetryRecoversFromPortAlreadyAllocated(t *testing.T) {
	createAndStart, calls := fakeCreateAndStart(errPortAllocated)
	containerId, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, true, testServiceUuid)
	require.NoError(t, err)
	require.Equal(t, "container-id", containerId)
	require.Equal(t, 2, *calls)
}

func TestUnrelatedErrorIsNotRetried(t *testing.T) {
	createAndStart, calls := fakeCreateAndStart(errUnrelated)
	_, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, true, testServiceUuid)
	require.Error(t, err)
	require.ErrorContains(t, err, "no such image")
	require.Equal(t, 1, *calls)
}

func TestCollisionWithoutAutoPublishedPortsIsNotRetried(t *testing.T) {
	// All host ports pinned by the user (NEAR static-port path): retrying a pinned
	// port can never succeed, so the collision must propagate immediately.
	createAndStart, calls := fakeCreateAndStart(errAddressInUse)
	_, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, false, testServiceUuid)
	require.Error(t, err)
	require.ErrorContains(t, err, "address already in use")
	require.Equal(t, 1, *calls)
}

func TestRetryGivesUpAfterMaxAttempts(t *testing.T) {
	persistentErrs := make([]error, maxHostPortBindRetries)
	for i := range persistentErrs {
		persistentErrs[i] = errAddressInUse
	}
	createAndStart, calls := fakeCreateAndStart(persistentErrs...)
	_, _, err := createAndStartContainerWithHostPortRetry(context.Background(), createAndStart, nil, true, testServiceUuid)
	require.Error(t, err)
	require.ErrorContains(t, err, "address already in use")
	require.Equal(t, maxHostPortBindRetries, *calls)
}

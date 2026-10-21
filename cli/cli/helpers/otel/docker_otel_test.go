package otel

import (
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/stretchr/testify/require"
)

func newClickHouseContainer(labels map[string]string) *types.Container {
	hostPortBindings := map[nat.Port]*nat.PortBinding{
		nat.Port(strconv.Itoa(int(clickHouseHTTPPort)) + "/tcp"): {HostIP: "", HostPort: strconv.Itoa(int(clickHouseHTTPHostPort))},
	}
	return types.NewContainer("id", "name", labels, types.ContainerStatus_Running, hostPortBindings, defaultClickHouseImage, nil, nil, nil, "")
}

func TestClickHouseContainerNeedsRecreationOnConfigHashMismatch(t *testing.T) {
	require.True(t, clickHouseContainerNeedsRecreation(newClickHouseContainer(map[string]string{})))
	require.True(t, clickHouseContainerNeedsRecreation(newClickHouseContainer(map[string]string{clickHouseConfigHashLabelKey.GetString(): "stale"})))
	require.False(t, clickHouseContainerNeedsRecreation(newClickHouseContainer(map[string]string{clickHouseConfigHashLabelKey.GetString(): clickHouseConfigHash})))
}

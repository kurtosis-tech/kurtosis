package docker_manager

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	labelSearchFilterKey             = "label"
	tinyTestImageNotAvailableOnArm64 = "clearlinux:base"
	arm64ArchitectureString          = "arm64"
)

func TestGetLabelsFilterList(t *testing.T) {
	//Enclave ID label
	enclaveKey := "enclaveID"
	enclaveID := "KTT2021-09-27T11.47.33-19414"

	//Container Type label
	containerTypeKey := "containerType"
	containerTypeValue := "service"

	labels := make(map[string]string)

	labels[enclaveKey] = enclaveID

	labelsFilterList := getLabelsFilterArgs(labelSearchFilterKey, labels)

	assert.False(t, labelsFilterList.MatchKVList(labelSearchFilterKey, nil))

	assert.True(t, labelsFilterList.MatchKVList(labelSearchFilterKey, map[string]string{
		enclaveKey: enclaveID,
	}))

	labels[containerTypeKey] = containerTypeValue

	labelsFilterList = getLabelsFilterArgs(labelSearchFilterKey, labels)

	assert.False(t, labelsFilterList.MatchKVList(labelSearchFilterKey, map[string]string{
		enclaveKey: enclaveID,
	}))

	assert.True(t, labelsFilterList.MatchKVList(labelSearchFilterKey, map[string]string{
		enclaveKey:       enclaveID,
		containerTypeKey: containerTypeValue,
	}))
}

func TestConvertCPUAllocationToNanoCPUsReturnsCorrectValue(t *testing.T) {
	cpuAllocation := uint64(1500)

	nanoCPUs := convertMillicpusToNanoCPUs(cpuAllocation)
	assert.Equal(t, uint64(1500000000), nanoCPUs)
}

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T) {
	memoryAllocationMegabytes := uint64(400) // 400 megabytes

	memoryAllocationBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
	assert.Equal(t, uint64(400000000), memoryAllocationBytes)
}

// We had a bug on 2022-09-19 where having IPv4 and IPv6 ports was incorrectly selecting the IPv6 one
// nolint: exhaustruct
func TestCorrectPortIsSelectedWhenIPv6IsPresent(t *testing.T) {
	dockerContainer := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:    "abc123",
			Name:  "noname",
			Image: "nginx",
			State: &container.State{
				Status: "running",
			},
		},
		NetworkSettings: &container.NetworkSettings{
			NetworkSettingsBase: container.NetworkSettingsBase{ //nolint:staticcheck // SA1019: NetworkSettingsBase is deprecated but still needed for Ports field
				Ports: nat.PortMap{
					"7443/tcp": []nat.PortBinding{
						{
							HostIP:   "::",
							HostPort: "49051",
						},
					},
					"7444/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "49050",
						},
					},
				},
			},
		},
		Config: &container.Config{
			Labels:     map[string]string{},
			Entrypoint: []string{},
			Cmd:        []string{},
			Env:        []string{},
		},
	}
	kurtosisContainer, err := newContainerFromDockerContainer(dockerContainer)
	require.NoError(t, err)

	hostPortBindings := kurtosisContainer.GetHostPortBindings()
	require.Equal(t, 1, len(hostPortBindings))

	portBinding, found := hostPortBindings["7444/tcp"]
	require.True(t, found)
	require.Equal(t, "127.0.0.1", portBinding.HostIP)
	require.Equal(t, "49050", portBinding.HostPort)
}

func TestCorrectSelectionWhenTwoOfSameIPs(t *testing.T) {
	port1 := nat.Port("9710/tcp")
	port2 := nat.Port("9711/tcp")

	// Directly from a case we saw in the field:
	// map[10000/tcp:[] 9710/tcp:[{HostIP:0.0.0.0 HostPort:9710}] 9711/tcp:[{HostIP:0.0.0.0 HostPort:9711}]]
	portMap := nat.PortMap{
		"10000/tcp": []nat.PortBinding{},
		"9710/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9710",
			},
		},
		"9711/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9711",
			},
		},
	}
	hostPortBindings := getHostPortBindingsOnExpectedInterface(portMap)
	require.Equal(t, 2, len(hostPortBindings))

	publicBinding1, found := hostPortBindings[port1]
	require.True(t, found)
	require.Equal(t, "127.0.0.1", publicBinding1.HostIP)
	require.Equal(t, "9710", publicBinding1.HostPort)

	publicBinding2, found := hostPortBindings[port2]
	require.True(t, found)
	require.Equal(t, "127.0.0.1", publicBinding2.HostIP)
	require.Equal(t, "9711", publicBinding2.HostPort)
}

func TestPullImageWithRetries(t *testing.T) {
	//if runtime.GOARCH != arm64ArchitectureString {
	//	t.Skip("Skipping the test as this is not running on arm64")
	//}
	//dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	//require.NoError(t, err)
	//require.NotNil(t, dockerClient)
	//ctx := context.Background()
	//err, retry := pullImage(ctx, dockerClient, tinyTestImageNotAvailableOnArm64, defaultPlatform)
	//require.Error(t, err)
	//require.True(t, retry)
	//err, retry = pullImage(ctx, dockerClient, tinyTestImageNotAvailableOnArm64, linuxAmd64)
	//require.NoError(t, err)
	//require.False(t, retry)
}

func TestBuildImage(t *testing.T) {
	//dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	//require.NoError(t, err)
	//require.NotNil(t, dockerClient)
	//
	//ctx := context.Background()
	//clientOpts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	//dockerManager, err := CreateDockerManager(clientOpts)
	//require.NoError(t, err)
	//
	//contextDirPath := ""
	//containerImageFilePath := contextDirPath + "/Dockerfile"
	//
	//imageBuildSpec := image_build_spec.NewImageBuildSpec(contextDirPath, containerImageFilePath, "")
	//_, err = dockerManager.BuildImage(ctx, "foobar", imageBuildSpec)
	//require.NoError(t, err)
}

func TestShmSizeDefaultsToZeroInArgsBuilder(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").Build()
	assert.Equal(t, uint64(0), args.shmSizeMegabytes)
}

func TestShmSizeIsStoredInArgsBuilder(t *testing.T) {
	const shmSizeMB = uint64(128)
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").
		WithShmSizeMegabytes(shmSizeMB).
		Build()
	assert.Equal(t, shmSizeMB, args.shmSizeMegabytes)
}

func TestShmSizeMegabytesToBytesConversion(t *testing.T) {
	// Docker HostConfig.ShmSize is in bytes; 128 MiB must equal 134217728 bytes.
	const shmSizeMB = uint64(128)
	expectedBytes := int64(134217728)
	assert.Equal(t, expectedBytes, int64(shmSizeMB)*shmMebibytesToBytesFactor)
}

func TestUlimitsDefaultToNilInArgsBuilder(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").Build()
	assert.Nil(t, args.ulimits)
}

func TestUlimitsAreStoredInArgsBuilder(t *testing.T) {
	ulimits := map[string]int64{"memlock": -1, "nofile": 65536}
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").
		WithUlimits(ulimits).
		Build()
	assert.Equal(t, ulimits, args.ulimits)
}

func TestBuildUlimitsReturnsNilForEmptyMap(t *testing.T) {
	result := buildUlimits(nil)
	assert.Nil(t, result)
	result = buildUlimits(map[string]int64{})
	assert.Nil(t, result)
}

func TestBuildUlimitsSetsHardAndSoftToSameValue(t *testing.T) {
	result := buildUlimits(map[string]int64{"memlock": -1})
	require.Len(t, result, 1)
	assert.Equal(t, "memlock", result[0].Name)
	assert.Equal(t, int64(-1), result[0].Soft)
	assert.Equal(t, int64(-1), result[0].Hard)
}

func TestGpuCountDefaultsToZeroInArgsBuilder(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").Build()
	assert.Equal(t, int64(0), args.gpuCount)
}

func TestGpuCountIsStoredInArgsBuilder(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("my-image", "my-container", "network-id").
		WithGpuCount(2).
		Build()
	assert.Equal(t, int64(2), args.gpuCount)
}

func TestBuildDeviceRequestsReturnsNilForZero(t *testing.T) {
	result := buildDeviceRequests(0, nil, "nvidia")
	assert.Nil(t, result)
}

func TestBuildDeviceRequestsForPositiveCount(t *testing.T) {
	result := buildDeviceRequests(2, nil, "nvidia")
	require.Len(t, result, 1)
	assert.Equal(t, "nvidia", result[0].Driver)
	assert.Equal(t, 2, result[0].Count)
	assert.Equal(t, [][]string{{"gpu"}}, result[0].Capabilities)
}

func TestBuildDeviceRequestsForAllGpus(t *testing.T) {
	result := buildDeviceRequests(-1, nil, "nvidia")
	require.Len(t, result, 1)
	assert.Equal(t, -1, result[0].Count)
}

package docker_manager

import (
	"testing"

	"github.com/docker/docker/api/types"
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
	dockerContainer := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    "abc123",
			Name:  "noname",
			Image: "nginx",
			State: &types.ContainerState{
				Status: "running",
			},
		},
		NetworkSettings: &types.NetworkSettings{
			NetworkSettingsBase: types.NetworkSettingsBase{
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

// func TestCommitAndSaveImage(t *testing.T) {
// 	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
// 	require.NoError(t, err)
// 	require.NotNil(t, dockerClient)

// 	ctx := context.Background()
// 	clientOpts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
// 	dockerManager, err := CreateDockerManager(clientOpts)
// 	require.NoError(t, err)

// 	containerId := "0a2c5a17a590"
// 	imageName := "tedim52/engine-snapshot:one"
// 	err = dockerManager.CommitContainer(ctx, containerId, imageName)
// 	require.NoError(t, err)
// 	logrus.Info("Successfully committed container to image")

// 	outputPath := "/Users/tewodrosmitiku/craft/kurtosis/engine-snapshot-two.tar"
// 	err = dockerManager.SaveImage(ctx, imageName, outputPath)
// 	require.NoError(t, err)
// 	logrus.Infof("Successfully saved image to %s\n", outputPath)

// 	err = dockerManager.LoadImage(ctx, outputPath)
// 	require.NoError(t, err)
// 	logrus.Info("Successfully loaded image from file")

// 	err = dockerManager.RemoveImage(ctx, imageName, false, false)
// 	require.NoError(t, err)
// 	logrus.Info("Successfully removed image")
// }

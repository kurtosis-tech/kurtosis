package docker_manager

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	labelSearchFilterKey = "label"
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
func TestCorrectPortIsSelectedWhenIPv6IsPresent(t *testing.T) {
	dockerContainer := types.Container{
		ID:      "abc123",
		Names:   []string{"noname"},
		Image:   "nginx",
		ImageID: "",
		Command: "",
		Created: 0,
		Ports: []types.Port{
			{
				IP:          "::",
				PrivatePort: 7443,
				PublicPort:  49051,
				Type:        "tcp",
			},
			{
				IP:          "0.0.0.0",
				PrivatePort: 7443,
				PublicPort:  49050,
				Type:        "tcp",
			},
		},
		SizeRw:     0,
		SizeRootFs: 0,
		Labels:     map[string]string{},
		State:      "running",
		Status:     "",
		HostConfig: struct {
			NetworkMode string `json:",omitempty"`
		}{},
		NetworkSettings: nil,
		Mounts:          nil,
	}
	kurtosisContainer, err := newContainerFromDockerContainer(dockerContainer)
	require.NoError(t, err)

	hostPortBindings := kurtosisContainer.GetHostPortBindings()
	require.Equal(t, 1, len(hostPortBindings))

	portBinding, found := hostPortBindings["7443/tcp"]
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

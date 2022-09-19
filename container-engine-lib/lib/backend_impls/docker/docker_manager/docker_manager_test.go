package docker_manager

import (
	"github.com/docker/docker/api/types"
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
		enclaveKey: enclaveID,
		containerTypeKey: containerTypeValue,
	}))
}

func TestConvertCPUAllocationToNanoCPUsReturnsCorrectValue(t *testing.T){
	cpuAllocation := uint64(1500)

	nanoCPUs := convertMillicpusToNanoCPUs(cpuAllocation)
	assert.Equal(t, uint64(1500000000), nanoCPUs)
}

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T){
	memoryAllocationMegabytes := uint64(400) // 400 megabytes

	memoryAllocationBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
	assert.Equal(t, uint64(400000000), memoryAllocationBytes)
}

// We had a bug on 2022-09-19 where having IPv4 and IPv6 ports was incorrectly selecting the IPv6 one
func TestCorrectPortIsSelectedWhenIPv6IsPresent(t *testing.T) {
	dockerContainer := types.Container{
		ID:         "abc123",
		Names:      []string{"noname"},
		Image:      "nginx",
		Ports:      []types.Port{
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
		Labels:     map[string]string{},
		State:     "running",
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
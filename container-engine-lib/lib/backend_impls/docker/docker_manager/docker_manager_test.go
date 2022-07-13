package docker_manager

import (
	"github.com/stretchr/testify/assert"
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
	cpuAllocationStr := "1.5"

	nanoCPUs, err := convertCPUAllocationToNanoCPUs(cpuAllocationStr)
	assert.NoError(t, err)

	assert.Equal(t, int64(1500000000), nanoCPUs)
}

func TestConvertCPUAllocationToNanoCPUsReturnsCorrectValueWithFractionLessThanZero(t *testing.T){
	cpuAllocationStr := "0.5"
	cpuAllocationStrNoZero := ".5"

	nanoCPUs, err := convertCPUAllocationToNanoCPUs(cpuAllocationStr)
	assert.NoError(t, err)
	nanoCPUsNoZero, err := convertCPUAllocationToNanoCPUs(cpuAllocationStrNoZero)
	assert.NoError(t, err)

	assert.Equal(t, int64(500000000), nanoCPUs)
	assert.Equal(t, int64(500000000), nanoCPUsNoZero)
}

func TestConvertCPUAllocationToNanoCPUsReturnsErrorWithInvalidFormat(t *testing.T){
	cpuAllocationStr := "one point five"

	_, err := convertCPUAllocationToNanoCPUs(cpuAllocationStr)
	assert.Error(t, err)
}

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T){
	memoryAllocation := uint64(400) // 400 megabytes

	memoryAllocationInBytes := convertMemoryAllocationToBytes(memoryAllocation)
	assert.Equal(t, uint64(400000000), memoryAllocationInBytes)
}
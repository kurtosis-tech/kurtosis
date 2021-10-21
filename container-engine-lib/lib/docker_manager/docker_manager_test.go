package docker_manager

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

	labelsFilterList := getLabelsFilterArgs(labels)

	assert.False(t, labelsFilterList.MatchKVList("label", nil))

	assert.True(t, labelsFilterList.MatchKVList("label", map[string]string{
		enclaveKey: enclaveID,
	}))

	labels[containerTypeKey] = containerTypeValue

	labelsFilterList = getLabelsFilterArgs(labels)

	assert.False(t, labelsFilterList.MatchKVList("label", map[string]string{
		enclaveKey: enclaveID,
	}))

	assert.True(t, labelsFilterList.MatchKVList("label", map[string]string{
		enclaveKey: enclaveID,
		containerTypeKey: containerTypeValue,
	}))
}

package user_services_functions

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T){
	memoryAllocationMegabytes := uint64(400) // 400 megabytes

	memoryAllocationBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
	require.Equal(t, uint64(400000000), memoryAllocationBytes)
}
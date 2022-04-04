/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/stretchr/testify/require"
	"net"
	"strings"
	"testing"
	"time"
)

const (
	packetLossPercentageValueForUnblockedPartitions float32 = 0
	packetLossPercentageValueForBlockedPartitions   float32 = 100
	packetLossPercentageValueForSoftPartition       float32 = 25

	testServiceGUID                  = "test"
	testEnclaveID              = "kt2022-03-17t16.33.01.495"
	testContainerStatusRunning = container_status.ContainerStatus_Running

	expectedCommandsForExecutingInitTrafficControl = "tc qdisc add dev eth1 root handle 1: htb && tc class add dev" +
		" eth1 parent 1: classid 1:1 htb rate 100% && tc class add dev eth1 parent 1: classid 1:2 htb rate 100% &&" +
		" tc filter add dev eth1 parent 1: handle 1:0 matchall flowid 1:1 && tc qdisc add dev eth1 parent 1:1 handle" +
		" 2: htb && tc qdisc add dev eth1 parent 1:2 handle 3: htb"

	expectedCommandsForExecutingBlockedPartitionInQdiscB = "tc qdisc del dev eth1 parent 1:2 handle 3: htb && tc qdisc " +
		"add dev eth1 parent 1:2 handle 3: htb && tc class add dev eth1 parent 3: classid 3:1 htb rate 100% && tc " +
		"filter add dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:1 match ip dst 1.1.1.1 && tc qdisc add dev " +
		"eth1 parent 3:1 handle 5: netem loss 100% && tc class add dev eth1 parent 3: classid 3:2 htb rate 100% && " +
		"tc filter add dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:2 match ip dst 2.2.2.2 && tc qdisc add " +
		"dev eth1 parent 3:2 handle 7: netem loss 100% && tc class add dev eth1 parent 3: classid 3:3 htb rate 100%" +
		" && tc filter add dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:3 match ip dst 3.3.3.3 && tc qdisc add" +
		" dev eth1 parent 3:3 handle 9: netem loss 100% && tc class add dev eth1 parent 3: classid 3:4 htb rate 100%" +
		" && tc filter add dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:4 match ip dst 4.4.4.4 && tc qdisc add" +
		" dev eth1 parent 3:4 handle b: netem loss 100% && tc filter replace dev eth1 parent 1: handle 1:0 matchall" +
		" flowid 1:2"

	expectedCommandsForExecutingSoftPartitionInQdiscA = "tc qdisc del dev eth1 parent 1:1 handle 2: htb && tc qdisc " +
		"add dev eth1 parent 1:1 handle 2: htb && tc class add dev eth1 parent 2: classid 2:1 htb rate 100% && tc " +
		"filter add dev eth1 parent 2: protocol ip prio 1 u32 flowid 2:1 match ip dst 1.1.1.1 && tc qdisc add dev " +
		"eth1 parent 2:1 handle 4: netem loss 25% && tc class add dev eth1 parent 2: classid 2:2 htb rate 100% && " +
		"tc filter add dev eth1 parent 2: protocol ip prio 1 u32 flowid 2:2 match ip dst 2.2.2.2 && tc qdisc add dev" +
		" eth1 parent 2:2 handle 6: netem loss 25% && tc class add dev eth1 parent 2: classid 2:3 htb rate 100% && tc" +
		" filter add dev eth1 parent 2: protocol ip prio 1 u32 flowid 2:3 match ip dst 3.3.3.3 && tc qdisc add dev " +
		"eth1 parent 2:3 handle 8: netem loss 25% && tc class add dev eth1 parent 2: classid 2:4 htb rate 100% && tc " +
		"filter add dev eth1 parent 2: protocol ip prio 1 u32 flowid 2:4 match ip dst 4.4.4.4 && tc qdisc add dev eth1" +
		" parent 2:4 handle a: netem loss 25% && tc filter replace dev eth1 parent 1: handle 1:0 matchall flowid 1:1"

	expectedCommandsForExecutingSoftPartitionInQdiscB = "tc qdisc del dev eth1 parent 1:2 handle 3: htb && tc qdisc add dev" +
		" eth1 parent 1:2 handle 3: htb && tc class add dev eth1 parent 3: classid 3:1 htb rate 100% && tc filter add" +
		" dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:1 match ip dst 1.1.1.1 && tc qdisc add dev eth1 parent" +
		" 3:1 handle 5: netem loss 25% && tc class add dev eth1 parent 3: classid 3:2 htb rate 100% && tc filter add" +
		" dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:2 match ip dst 2.2.2.2 && tc qdisc add dev eth1 parent" +
		" 3:2 handle 7: netem loss 25% && tc class add dev eth1 parent 3: classid 3:3 htb rate 100% && tc filter add" +
		" dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:3 match ip dst 3.3.3.3 && tc qdisc add dev eth1 parent " +
		"3:3 handle 9: netem loss 25% && tc class add dev eth1 parent 3: classid 3:4 htb rate 100% && tc filter add " +
		"dev eth1 parent 3: protocol ip prio 1 u32 flowid 3:4 match ip dst 4.4.4.4 && tc qdisc add dev eth1 parent 3:4" +
		" handle b: netem loss 25% && tc filter replace dev eth1 parent 1: handle 1:0 matchall flowid 1:2"

	expectedCommandsForExecutingUnblockedPartition = "tc qdisc del dev eth1 parent 1:1 handle 2: htb && tc qdisc del" +
		" dev eth1 parent 1:2 handle 3: htb && tc qdisc add dev eth1 parent 1:1 handle 2: htb && tc qdisc add dev " +
		"eth1 parent 1:2 handle 3: htb"

	stringSeparatorInCommand = " "
)

var (
	testNetworkinSidecarIP = []byte{1, 2, 3, 4}

	userServiceTest1IPAddress     = net.ParseIP("1.1.1.1")
	userServiceTest2IPAddress     = net.ParseIP("2.2.2.2")
	userServiceTest3IPAddress     = net.ParseIP("3.3.3.3")
	userServiceTest4IPAddress     = net.ParseIP("4.4.4.4")
	allUserServiceTestIPAddresses = []net.IP{
		userServiceTest1IPAddress, userServiceTest2IPAddress, userServiceTest3IPAddress, userServiceTest4IPAddress,
	}
	//the value on this map represents the key in allUserServiceTestIPAddresses
	allUserServiceTestIPAddressesMap = map[string]int{
		userServiceTest1IPAddress.String(): 0,
		userServiceTest2IPAddress.String(): 1,
		userServiceTest3IPAddress.String(): 2,
		userServiceTest4IPAddress.String(): 3,
	}

	qdiscAChildrenQdiscId                 = []qdiscID{qdiscID("4:"), qdiscID("6:"), qdiscID("8:"), qdiscID("a:"), qdiscID("c:"), qdiscID("e:"), qdiscID("10:")}
	qdiscAChildrenQdiscDecimalMajorNumber = []int{4, 6, 8, 10, 12, 14, 16}
	qdiscBChildrenQdiscId                 = []qdiscID{qdiscID("5:"), qdiscID("7:"), qdiscID("9:"), qdiscID("b:"), qdiscID("d:"), qdiscID("f:"), qdiscID("11:")}
	qdiscBChildrenDecimalMajorNumber      = []int{5, 7, 9, 11, 13, 15, 17}

	qdiscAChildrenClassId                 = []classID{classID("2:1"), classID("2:2"), classID("2:3"), classID("2:4")}
	qdiscAChildrenClassDecimalMinorNumber = []int{1, 2, 3, 4}
	qdiscBChildrenClassId                 = []classID{classID("3:1"), classID("3:2"), classID("3:3"), classID("3:4")}
	qdiscBChildrenClassDecimalMinorNumber = []int{1, 2, 3, 4}
)

func TestInitializeTrafficControl(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, execCmdExecutor := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)

	err := sidecar.InitializeTrafficControl(ctx)
	require.NoError(t, err, "An error occurred initializing traffic control")
	require.Equal(t, initialKurtosisQdiscId, sidecar.qdiscInUse)
	require.Equal(t, 1, len(execCmdExecutor.commands))

	actualFirstExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[0])
	require.Equal(t, expectedCommandsForExecutingInitTrafficControl, actualFirstExecutedMergedCmd)
}

func TestInitializeTrafficControl_AlreadyInitialized(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, _ := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	sidecar.qdiscInUse = initialKurtosisQdiscId

	err := sidecar.InitializeTrafficControl(ctx)
	require.Nil(t, err, "Traffic control already initialized")
}

func TestUpdateTrafficControl_CreateBlockedPartitionAndThenUnblockIt(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, execCmdExecutor := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)
	sidecar.qdiscInUse = initialKurtosisQdiscId

	//Blocking partition
	allUserServicePacketLossConfigurationsForBlockedPartition := getAllUserServicePacketLossConfigurationsForBlockedPartition()

	err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForBlockedPartition)
	require.NoError(t, err, "An error occurred updating qdisc configuration for blocked partition")
	require.Equal(t, 1, len(execCmdExecutor.commands))

	actualFirstExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[0])
	require.Equal(t, expectedCommandsForExecutingBlockedPartitionInQdiscB, actualFirstExecutedMergedCmd)

	//Unblocking partition
	allUserServicePacketLossConfigurationsForUnblockedPartition := getAllUserServicePacketLossConfigurationsForUnblockedPartition()

	err = sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForUnblockedPartition)
	require.NoError(t, err, "An error occurred updating qdisc configuration for unblocked partition")
	require.Equal(t, initialKurtosisQdiscId, sidecar.qdiscInUse)
	require.Equal(t, 2, len(execCmdExecutor.commands))

	actualSecondExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[1])
	require.Equal(t, expectedCommandsForExecutingUnblockedPartition, actualSecondExecutedMergedCmd)
}

func TestUpdateTrafficControl_CreateSoftPartitionAndThenUnblockIt(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, execCmdExecutor := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)
	sidecar.qdiscInUse = initialKurtosisQdiscId

	//Soft partition
	allUserServicePacketLossConfigurations := getAllUserServicePacketLossConfigurationsForSoftPartition()

	err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurations)
	require.NoError(t, err, "An error occurred updating qdisc configuration for soft partition")
	require.Equal(t, 1, len(execCmdExecutor.commands))

	actualFirstExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[0])
	require.Equal(t, expectedCommandsForExecutingSoftPartitionInQdiscB, actualFirstExecutedMergedCmd)

	//Unblocking partition
	allUserServicePacketLossConfigurationsForUnblockedPartition := getAllUserServicePacketLossConfigurationsForUnblockedPartition()

	err = sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForUnblockedPartition)
	require.NoError(t, err, "An error occurred updating qdisc configuration for unblocked partition")
	require.Equal(t, initialKurtosisQdiscId, sidecar.qdiscInUse)
	require.Equal(t, 2, len(execCmdExecutor.commands))

	actualSecondExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[1])
	require.Equal(t, expectedCommandsForExecutingUnblockedPartition, actualSecondExecutedMergedCmd)
}

func TestUpdateTrafficControl_CreateBlockedPartitionAndThenSoftPartition(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, execCmdExecutor := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)
	sidecar.qdiscInUse = initialKurtosisQdiscId

	//Blocking partition
	allUserServicePacketLossConfigurationsForBlockedPartition := getAllUserServicePacketLossConfigurationsForBlockedPartition()

	err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForBlockedPartition)
	require.Equal(t, qdiscBID, sidecar.qdiscInUse)
	require.NoError(t, err, "An error occurred updating qdisc configuration for blocked partition")
	require.Equal(t, 1, len(execCmdExecutor.commands))

	actualFirstExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[0])
	require.Equal(t, expectedCommandsForExecutingBlockedPartitionInQdiscB, actualFirstExecutedMergedCmd)

	//Unblocking partition
	allUserServicePacketLossConfigurationsForSoftPartition := getAllUserServicePacketLossConfigurationsForSoftPartition()

	err = sidecar.UpdateTrafficControl(context.Background(), allUserServicePacketLossConfigurationsForSoftPartition)
	require.NoError(t, err, "An error occurred updating qdisc configuration for soft partition")
	require.Equal(t, qdiscAID, sidecar.qdiscInUse)
	require.Equal(t, 2, len(execCmdExecutor.commands))

	actualSecondExecutedMergedCmd := mergeCommandsInOneLine(execCmdExecutor.commands[1])
	require.Equal(t, expectedCommandsForExecutingSoftPartitionInQdiscA, actualSecondExecutedMergedCmd)
}

func TestUpdateTrafficControl_UndefinedQdiscInUseError(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, _ := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)

	//Execution
	allUserServicePacketLossConfigurationsForBlockedPartition := getAllUserServicePacketLossConfigurationsForBlockedPartition()
	err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForBlockedPartition)
	require.Error(t, err, "Expected undefined qdisc id in use error")
}

func TestUpdateTrafficControl_UnrecognizedPrimaryQdiscIdError(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, _ := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)
	sidecar.qdiscInUse = "1:"

	//Execution
	allUserServicePacketLossConfigurationsForBlockedPartition := getAllUserServicePacketLossConfigurationsForBlockedPartition()
	err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurationsForBlockedPartition)
	require.Error(t, err, "Expected unrecognized primary qdisc id error")
}

func TestGetNextUnusedQdiscId_GenereratQdiscAChildren(t *testing.T) {

	parentQdiscID := qdiscAID
	previousQdiscIdDecimalMajorNumber := lastUsedQdiscIdDecimalMajorNumber

	for childKey, expectedChildQdiscId := range qdiscAChildrenQdiscId {
		nextQdiscID, decimalMajorNumber, err := getNextUnusedQdiscId(parentQdiscID, previousQdiscIdDecimalMajorNumber)
		previousQdiscIdDecimalMajorNumber = decimalMajorNumber
		require.NoError(t, err, "An error occurred creating next unused qdisc id")
		expectedDecimalMajorNumber := qdiscAChildrenQdiscDecimalMajorNumber[childKey]
		require.Equal(t, expectedChildQdiscId, nextQdiscID)
		require.Equal(t, expectedDecimalMajorNumber, decimalMajorNumber)
	}
}

func TestGetNextUnusedQdiscId_GenereratQdiscBChildren(t *testing.T) {

	parentQdiscID := qdiscBID
	previousQdiscIdDecimalMajorNumber := lastUsedQdiscIdDecimalMajorNumber

	for childKey, expectedChildQdiscId := range qdiscBChildrenQdiscId {
		nextQdiscID, decimalMajorNumber, err := getNextUnusedQdiscId(parentQdiscID, previousQdiscIdDecimalMajorNumber)
		previousQdiscIdDecimalMajorNumber = decimalMajorNumber
		require.NoError(t, err, "An error occurred creating next unused qdisc id")
		expectedDecimalMajorNumber := qdiscBChildrenDecimalMajorNumber[childKey]
		require.Equal(t, expectedChildQdiscId, nextQdiscID)
		require.Equal(t, expectedDecimalMajorNumber, decimalMajorNumber)
	}
}

func TestGetNextUnusedQdiscId_UnrecognizedParentQdiscIdError(t *testing.T) {
	parentQdiscID := qdiscID("1:")
	previousQdiscIdDecimalMajorNumber := lastUsedQdiscIdDecimalMajorNumber

	_, _, err := getNextUnusedQdiscId(parentQdiscID, previousQdiscIdDecimalMajorNumber)
	require.Error(t, err, "Return an error because unrecognized parent qdisc id")
}

func TestNewClassId_QdiscAChildren(t *testing.T) {

	parentQdiscId := qdiscAID

	for childKey, expectedChildClassId  := range qdiscAChildrenClassId {
		decimalMinorNumber := qdiscAChildrenClassDecimalMinorNumber[childKey]
		actualClassId := newClassId(parentQdiscId, decimalMinorNumber)
		require.Equal(t, expectedChildClassId, actualClassId)
	}
}

func TestNewClassId_QdiscBChildren(t *testing.T) {

	parentQdiscId := qdiscBID

	for childKey, expectedChildClassId  := range qdiscBChildrenClassId {
		decimalMinorNumber := qdiscBChildrenClassDecimalMinorNumber[childKey]
		actualClassId := newClassId(parentQdiscId, decimalMinorNumber)
		require.Equal(t, expectedChildClassId, actualClassId)
	}
}

func TestConcurrencySafety(t *testing.T) {
	//Initial state
	ctx := context.Background()
	sidecar, execCmdExecutor := createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t)
	require.Empty(t, sidecar.qdiscInUse)

	numProcesses := 2

	err := sidecar.InitializeTrafficControl(ctx)

	require.NoErrorf(t, err, "An error occurred initiliazing traffic control")

	execCmdExecutor.setBlocked(true)

	for i := 0; i < numProcesses; i++ {
		iByte := byte(i)
		ip := net.IP{iByte, iByte, iByte, iByte}
		allUserServicePacketLossConfigurations := map[string]float32{}
		allUserServicePacketLossConfigurations[ip.String()] = packetLossPercentageValueForBlockedPartitions
		go func() {
			err := sidecar.UpdateTrafficControl(ctx, allUserServicePacketLossConfigurations)
			require.NoErrorf(t, err, "An error occurred updating traffic control")
		}()
		time.Sleep(5 * time.Millisecond)  // Make sure they enter the sidecar in proper order
	}

	// At this point:
	// - If the sidecar isn't controlling concurrency, all the processes will be backed up inside the exec cmd executor
	// - If the sidecar is controlling concurrency, only one thread will be in the ExecCmdExecutor and the rest will be queued
	//     inside the sidecar in FIFO order

	execCmdExecutor.setBlocked(false)

	// Give the now-unblocked threads time to finish
	time.Sleep(500 * time.Millisecond)

	// Verify that concurrency was controlled in the sidecar, so everything is ordered
	// We ignore the first command, because it will be the initialization
	for i := 1; i <= numProcesses; i++ {
		expectedByte := byte(i - 1)
		expectedIpStr := net.IP([]byte{expectedByte, expectedByte, expectedByte, expectedByte}).String()
		expectedMatchIpDst := fmt.Sprintf("match ip dst %v", expectedIpStr )
		actualCmd := strings.Join(execCmdExecutor.commands[i], stringSeparatorInCommand)
		require.Contains(t, actualCmd, expectedMatchIpDst)
	}
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getAllUserServicePacketLossConfigurationsForSoftPartition() map[string]float32 {
	allUserServicePacketLossConfigurations := map[string]float32{}
	for _, ip := range allUserServiceTestIPAddresses {
		allUserServicePacketLossConfigurations[ip.String()] = packetLossPercentageValueForSoftPartition
	}
	return allUserServicePacketLossConfigurations
}

func getAllUserServicePacketLossConfigurationsForBlockedPartition() map[string]float32 {
	allUserServicePacketLossConfigurations := map[string]float32{}
	for _, ip := range allUserServiceTestIPAddresses {
		allUserServicePacketLossConfigurations[ip.String()] = packetLossPercentageValueForBlockedPartitions
	}
	return allUserServicePacketLossConfigurations
}

func getAllUserServicePacketLossConfigurationsForUnblockedPartition() map[string]float32 {
	allUserServicePacketLossConfigurations := map[string]float32{}
	for _, ip := range allUserServiceTestIPAddresses {
		allUserServicePacketLossConfigurations[ip.String()] = packetLossPercentageValueForUnblockedPartitions
	}
	return allUserServicePacketLossConfigurations
}

func createNewStandardNetworkingSidecarAndMockedExecCmdExecutor(t *testing.T) (*StandardNetworkingSidecarWrapper, *mockSidecarExecCmdExecutor) {
	execCmdExecutor := newMockSidecarExecCmdExecutor()

	networkingSidecar := networking_sidecar.NewNetworkingSidecar(
		testServiceGUID,
		testNetworkinSidecarIP,
		testEnclaveID,
		testContainerStatusRunning)

	sidecar, err := NewStandardNetworkingSidecarWrapper(
		networkingSidecar,
		execCmdExecutor,
		)

	require.NoErrorf(t, err, "An error occurred creating standard networking sidecar wrapper with mocked exec command executor")

	return sidecar, execCmdExecutor
}

func mergeCommandsInOneLine(commands []string) string {
	//First order ip addresses in the sentence.
	allUserServiceTestIPAddressesKey := 0
	for cmdKey, cmd := range commands {
		var ip net.IP
		_, found := allUserServiceTestIPAddressesMap[cmd]
		if found {
			ip = allUserServiceTestIPAddresses[allUserServiceTestIPAddressesKey]
			commands[cmdKey] = ip.String()
			allUserServiceTestIPAddressesKey++
		}
	}
	result := strings.Join(commands, stringSeparatorInCommand)

	return result
}

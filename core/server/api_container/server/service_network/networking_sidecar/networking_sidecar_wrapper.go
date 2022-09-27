/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

const (
	tcCommand                   = "tc"
	tcAddCommand                = "add"
	tcReplaceCommand            = "replace"
	tcDeleteCommand             = "del"
	tcQdiscCommand              = "qdisc"
	tcQdiscTypeHtb              = "htb"
	tcQdiscTypeNetem            = "netem"
	tcQdiscTypeNetemOptionLoss  = "loss"
	tcClassCommand              = "class"
	tcFilterCommand             = "filter"
	tcFilterProtocolCommand     = "protocol"
	tcFilterIPCommand           = "ip"
	tcFilterPrioCommand         = "prio"
	tcFilterFlowIDCommand       = "flowid"
	tcFilterMatchCommand        = "match"
	tcFilterBasicTypeCommand    = "basic"
	tcFilterIPMatchTypeCommand  = "ip"
	tcFilterIPDestCommand       = "dst"
	tcU32FilterTypeCommand      = "u32"
	tcDeviceCommand             = "dev"
	tcHandleCommand             = "handle"
	tcParentCommand             = "parent"
	tcClassIDCommand            = "classid"
	tcRateCommand               = "rate"

	rootQdiscName                 = "root"
	defaultDockerNetworkInterface = "eth1"

	rootQdiscID                       qdiscID = "1:"
	qdiscAID                          qdiscID = "2:"
	qdiscBID                          qdiscID = "3:"
	lastUsedQdiscIdDecimalMajorNumber         = 3
	undefinedQdiscId                          = ""
	initialKurtosisQdiscId                    = qdiscAID

	rootFilterID                   = filterID(rootQdiscID) + "0"
	rootClassAClassID              = classID(rootQdiscID) + "1"
	rootClassBClassID              = classID(rootQdiscID) + "2"
	fullRateValue                  = "100%"
	maxFilterPriority              = "1"
	percentageSign                 = "%"
	firstClassIdDecimalMinorNumber = 1

	concatenateCommandsOperator = "&&"

	firstCommandIndex = 0

	unblockedPartitionConnectionPacketLossPercentageValue float32 = 0
)

// ==========================================================================================
//                                        Interface
// ==========================================================================================
// Extracted as interface for testing
type NetworkingSidecarWrapper interface {
	GetServiceGUID() service.ServiceGUID
	GetIPAddr() net.IP
	InitializeTrafficControl(ctx context.Context) error
	UpdateTrafficControl(ctx context.Context, allPacketLossPercentageForIpAddresses map[string]float32) error
}

// ==========================================================================================
//                                      Implementation
// ==========================================================================================
type qdiscID string
type classID string
type filterID string

// Provides a handle into manipulating the network state of a service container indirectly, via the sidecar
type StandardNetworkingSidecarWrapper struct {
	mutex *sync.Mutex

	networkingSidecar *networking_sidecar.NetworkingSidecar

	sidecarIpAddr net.IP
	// Tracks which of the main qdiscs (qdiscA and qdiscB) is the primary qdisc, so we know
	//  which qdisc is in the background that we can flush and rebuild
	//  when we're changing them
	qdiscInUse qdiscID

	execCmdExecutor sidecarExecCmdExecutor
}

func NewStandardNetworkingSidecarWrapper(
	networkingSidecar *networking_sidecar.NetworkingSidecar,
	execCmdExecutor sidecarExecCmdExecutor,
) (
	*StandardNetworkingSidecarWrapper,
	error,
) {
	if networkingSidecar == nil {
		return nil, stacktrace.NewError("The networking sidecar parameter must not be nil")
	}

	return &StandardNetworkingSidecarWrapper{
		mutex:             &sync.Mutex{},
		networkingSidecar: networkingSidecar,
		qdiscInUse:        undefinedQdiscId,
		execCmdExecutor:   execCmdExecutor,
	}, nil
}

func (sidecarWrapper *StandardNetworkingSidecarWrapper) GetServiceGUID() service.ServiceGUID {
	return sidecarWrapper.networkingSidecar.GetServiceGUID()
}

func (sidecarWrapper *StandardNetworkingSidecarWrapper) GetIPAddr() net.IP {
	return sidecarWrapper.sidecarIpAddr
}

func (sidecarWrapper *StandardNetworkingSidecarWrapper) InitializeTrafficControl(ctx context.Context) error {
	sidecarWrapper.mutex.Lock()
	defer sidecarWrapper.mutex.Unlock()
	if sidecarWrapper.qdiscInUse != undefinedQdiscId {
		return nil
	}

	initCmd := generateTcInitCmd()

	cmdDescription := "tc init"

	if err := sidecarWrapper.executeCmdInSidecar(ctx, initCmd, cmdDescription); err != nil {
		return stacktrace.Propagate(err, "An error occurred executing cmd '%v' in networking sidecar with GUID '%v'", initCmd, sidecarWrapper.GetServiceGUID())
	}

	sidecarWrapper.qdiscInUse = initialKurtosisQdiscId

	return nil
}

func (sidecarWrapper *StandardNetworkingSidecarWrapper) UpdateTrafficControl(ctx context.Context, allPacketLossPercentageForIpAddresses map[string]float32) error {
	sidecarWrapper.mutex.Lock()
	defer sidecarWrapper.mutex.Unlock()

	if sidecarWrapper.qdiscInUse == undefinedQdiscId {
		return stacktrace.NewError("Cannot update tc qdiscs because they haven't yet been initialized")
	}

	var isAnyPartitionBlocked bool
	for _, packetLossPercentage := range allPacketLossPercentageForIpAddresses {
		if packetLossPercentage > unblockedPartitionConnectionPacketLossPercentageValue {
			isAnyPartitionBlocked = true
		}
	}

	if isAnyPartitionBlocked && len(allPacketLossPercentageForIpAddresses) > 0 {
		primaryQdisc := sidecarWrapper.qdiscInUse
		var backgroundQdisc qdiscID
		var backgroundQdiscClass classID
		if primaryQdisc == qdiscAID {
			backgroundQdisc = qdiscBID
			backgroundQdiscClass = rootClassBClassID
		} else if primaryQdisc == qdiscBID {
			backgroundQdisc = qdiscAID
			backgroundQdiscClass = rootClassAClassID
		} else {
			return stacktrace.NewError("Unrecognized tc qdisc ID '%v' in use; this is a code bug", primaryQdisc)
		}

		updateTcCmd, err := generateTcUpdateCmd(backgroundQdisc, backgroundQdiscClass, allPacketLossPercentageForIpAddresses)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred generating tc update command with background qdisc ID '%v', "+
					"background qdisc class ID '%v' and all packet loss percentage for IPs %+v ", backgroundQdisc, backgroundQdiscClass, allPacketLossPercentageForIpAddresses)
		}

		cmdDescription := "tc update"

		if err := sidecarWrapper.executeCmdInSidecar(ctx, updateTcCmd, cmdDescription); err != nil {
			return stacktrace.Propagate(err, "An error occurred executing cmd '%v' inside the sidecar container", cmdDescription)
		}

		sidecarWrapper.qdiscInUse = backgroundQdisc
	} else if !isAnyPartitionBlocked {
		//if isAnyPartitionBlocked == false means the tc qdisc config has to be re-initialized (e.g.: when an unblocked partition is configured).
		//This is going to be done deleting and recreating qdiscA and qdiscB
		reInitQdiscAAndQdiscBCmd := generateTcReInitQdiscAAndQdiscBCmd()

		cmdDescription := "tc re init qdisc A and qdisc B cmd"

		if err := sidecarWrapper.executeCmdInSidecar(ctx, reInitQdiscAAndQdiscBCmd, cmdDescription); err != nil {
			return stacktrace.Propagate(err, "An error occurred executing cmd '%v' inside the sidecar container", cmdDescription)
		}

		sidecarWrapper.qdiscInUse = initialKurtosisQdiscId
	}

	return nil
}

// ==========================================================================================
//                                   Private helper functions
// ==========================================================================================
func getNextUnusedQdiscId(parentQdisc qdiscID, previousQdiscIdDecimalMajorNumber int) (qdiscID, int, error) {
	//This func receives the most-recently-created qdisc ID major number (in decimal, i.e. base-10),
	//and returns the ID (in hex, i.e. base-16) of the next qdisc that should be created.
	//The function works by finding the next even or odd ID number after the most-recently-created
	//qdisc ID - even if the qdisc-to-be-created's parent is qdisc A, and odd if the qdisc-to-be-created's
	//parent is qdisc B.
	decimalMajorNumber := previousQdiscIdDecimalMajorNumber + 1
	if parentQdisc == qdiscAID { //Qdisc A children should have even qdiscIds
		if !isEvenNumber(decimalMajorNumber) {
			decimalMajorNumber++
		}
	} else if parentQdisc == qdiscBID { //Qdisc B children should have odd qdiscIds
		if isEvenNumber(decimalMajorNumber) {
			decimalMajorNumber++
		}
	} else {
		return "", decimalMajorNumber, stacktrace.NewError("Unrecognized tc qdisc ID '%v' in use; this is a code bug", parentQdisc)
	}
	qdiscIdStr := fmt.Sprintf("%x:", decimalMajorNumber)
	return qdiscID(qdiscIdStr), decimalMajorNumber, nil
}

func newClassId(parentQdisc qdiscID, decimalMinorNumber int) classID {
	classIdStr := fmt.Sprintf("%v%x", parentQdisc, decimalMinorNumber)
	return classID(classIdStr)
}

func isEvenNumber(number int) bool {
	return number%2 == 0
}

func generateTcInitCmd() []string {
	commandList := [][]string{
		generateTcAddRootQdiscCmd(),
		generateTcAddClassACmd(),
		generateTcAddClassBCmd(),
		generateTcAddRootFilterCmd(),
		generateTcAddQdiscACmd(),
		generateTcAddQdiscBCmd(),
	}

	resultCmd := mergeCommandListInOneLineCommand(commandList)

	return resultCmd
}

func generateTcUpdateCmd(backgroundQdisc qdiscID, backgroundQdiscClass classID, allPacketLossPercentageForIpAddresses map[string]float32) ([]string, error) {
	commandList := [][]string{
		generateTcRemoveQdiscCmd(backgroundQdiscClass, backgroundQdisc),              //First remove all background Qdisc configuration in order to recreate it
		generateTcAddQdiscCmd(backgroundQdiscClass, backgroundQdisc, tcQdiscTypeHtb), //Creating the background Qdisc again to fill it with new configuration
	}

	parentQdisc := backgroundQdisc
	classIdDecimalMinorNumber := firstClassIdDecimalMinorNumber
	previousQdiscIdDecimalMajorNumber := lastUsedQdiscIdDecimalMajorNumber
	for ipAddress, packetLossPercentage := range allPacketLossPercentageForIpAddresses {
		classId := newClassId(parentQdisc, classIdDecimalMinorNumber)
		classIdDecimalMinorNumber++
		qdiscId, decimalMajorNumber, err := getNextUnusedQdiscId(parentQdisc, previousQdiscIdDecimalMajorNumber)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new qdisc ID for parent qdisc ID %v and previous qdisc ID decimal major number %v", parentQdisc, previousQdiscIdDecimalMajorNumber)
		}
		previousQdiscIdDecimalMajorNumber = decimalMajorNumber
		//For each ip address that will be blocked, we create: a class that will be a child of QdiscA or QdiscB, a filter
		//pointing to this class, and a netem qdisc (which is a child of the class) with the packet loss configuration
		commandList = append(commandList, generateTcAddClassCmd(parentQdisc, classId))
		commandList = append(commandList, generateTCAddFilterByIpCmd(parentQdisc, classId, ipAddress))
		commandList = append(commandList, generateTCAddNetemQdiscWithPacketLossCmd(classId, qdiscId, packetLossPercentage))
	}

	commandList = append(commandList, generateTCReplaceRootFilterCmd(backgroundQdiscClass)) //swapping the root filter pointer

	resultCmd := mergeCommandListInOneLineCommand(commandList)

	return resultCmd, nil
}

func generateTcReInitQdiscAAndQdiscBCmd() []string {
	commandList := [][]string{
		generateTcRemoveQdiscACmd(),
		generateTcRemoveQdiscBCmd(),
		generateTcAddQdiscACmd(),
		generateTcAddQdiscBCmd(),
	}

	resultCmd := mergeCommandListInOneLineCommand(commandList)

	return resultCmd
}

func generateTcAddRootQdiscCmd() []string {
	resultCmd := []string{
		tcCommand,
		tcQdiscCommand,
		tcAddCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		rootQdiscName,
		tcHandleCommand,
		string(rootQdiscID),
		tcQdiscTypeHtb,
	}

	return resultCmd
}

func generateTcAddClassACmd() []string {
	return generateTcAddClassCmd(rootQdiscID, rootClassAClassID)
}

func generateTcAddClassBCmd() []string {
	return generateTcAddClassCmd(rootQdiscID, rootClassBClassID)
}

func generateTcAddRootFilterCmd() []string {

	resultCmd := []string{
		tcCommand,
		tcFilterCommand,
		tcAddCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(rootQdiscID),
		tcHandleCommand,
		string(rootFilterID),
		tcFilterBasicTypeCommand,
		tcFilterFlowIDCommand,
		string(rootClassAClassID),
	}

	return resultCmd
}

func generateTcAddQdiscACmd() []string {
	return generateTcAddQdiscCmd(rootClassAClassID, qdiscAID, tcQdiscTypeHtb)
}

func generateTcRemoveQdiscACmd() []string {
	return generateTcRemoveQdiscCmd(rootClassAClassID, qdiscAID)
}

func generateTcRemoveQdiscBCmd() []string {
	return generateTcRemoveQdiscCmd(rootClassBClassID, qdiscBID)
}

func generateTcAddQdiscBCmd() []string {
	return generateTcAddQdiscCmd(rootClassBClassID, qdiscBID, tcQdiscTypeHtb)
}

func generateTCAddNetemQdiscWithPacketLossCmd(parentClassId classID, qdiscId qdiscID, packetLossPercentage float32) []string {

	packetLossPercentageStr := fmt.Sprintf("%v%v", packetLossPercentage, percentageSign)

	resultCmd := generateTcAddQdiscCmd(parentClassId, qdiscId, tcQdiscTypeNetem)
	resultCmd = append(resultCmd, tcQdiscTypeNetemOptionLoss)
	resultCmd = append(resultCmd, packetLossPercentageStr)

	return resultCmd
}

func generateTcAddQdiscCmd(parentClassId classID, qdiscId qdiscID, qdiscType string) []string {
	resultCmd := []string{
		tcCommand,
		tcQdiscCommand,
		tcAddCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(parentClassId),
		tcHandleCommand,
		string(qdiscId),
		qdiscType,
	}

	return resultCmd
}

func generateTcAddClassCmd(parentQdiscId qdiscID, classId classID) []string {

	resultCmd := []string{
		tcCommand,
		tcClassCommand,
		tcAddCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(parentQdiscId),
		tcClassIDCommand,
		string(classId),
		tcQdiscTypeHtb,
		tcRateCommand,
		fullRateValue,
	}

	return resultCmd
}

func generateTCAddFilterByIpCmd(parentQdiscId qdiscID, classId classID, ipAddress string) []string {

	resultCmd := []string{
		tcCommand,
		tcFilterCommand,
		tcAddCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(parentQdiscId),
		tcFilterProtocolCommand,
		tcFilterIPCommand,
		tcFilterPrioCommand,
		maxFilterPriority,
		tcU32FilterTypeCommand,
		tcFilterFlowIDCommand,
		string(classId),
		tcFilterMatchCommand,
		tcFilterIPMatchTypeCommand,
		tcFilterIPDestCommand,
		ipAddress,
	}

	return resultCmd
}

func generateTcRemoveQdiscCmd(parentClassId classID, qdiscId qdiscID) []string {
	resultCmd := []string{
		tcCommand,
		tcQdiscCommand,
		tcDeleteCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(parentClassId),
		tcHandleCommand,
		string(qdiscId),
		tcQdiscTypeHtb,
	}

	return resultCmd
}

func generateTCReplaceRootFilterCmd(classId classID) []string {

	resultCmd := []string{
		tcCommand,
		tcFilterCommand,
		tcReplaceCommand,
		tcDeviceCommand,
		defaultDockerNetworkInterface,
		tcParentCommand,
		string(rootQdiscID),
		tcHandleCommand,
		string(rootFilterID),
		tcFilterBasicTypeCommand,
		tcFilterFlowIDCommand,
		string(classId),
	}

	return resultCmd
}

func mergeCommandListInOneLineCommand(commandList [][]string) []string {
	resultCmd := []string{}
	for commandIndex, command := range commandList {
		if commandIndex > firstCommandIndex {
			resultCmd = append(resultCmd, concatenateCommandsOperator)
		}
		resultCmd = append(resultCmd, command...)
	}

	return resultCmd
}

func (sidecarWrapper *StandardNetworkingSidecarWrapper) executeCmdInSidecar(ctx context.Context, cmd []string, cmdDescription string) error {

	logrus.Infof(
		"Running %v command '%+v' in networking sidecar with service GUID '%v'...",
		cmdDescription,
		cmd,
		sidecarWrapper.networkingSidecar.GetServiceGUID())

	if err := sidecarWrapper.execCmdExecutor.exec(ctx, cmd); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running %v command",
			cmdDescription)
	}

	logrus.Infof("Successfully executed %v command against networking sidecar with GUID '%v'", cmdDescription, sidecarWrapper.GetServiceGUID())

	return nil
}

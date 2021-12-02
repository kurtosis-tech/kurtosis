/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_container_launcher

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

// ====================================================================================================
//                                    Get Port Maps Tests
// ====================================================================================================
func TestNormalPortMapRetrieval(t *testing.T) {
	port1Id := "port1"
	port1Num := uint16(1234)
	port1Proto := EnclaveContainerPortProtocol_TCP
	expectedPort1DockerSpec := nat.Port(fmt.Sprintf(
		"%v/%v",
		port1Num,
		port1Proto,
	))
	expectedPort1EnclaveContainerPort, err := NewEnclaveContainerPort(port1Num, port1Proto)
	require.NoError(t, err, "Unexpected error occurred building the port 1 spec")

	port2Id := "port2"
	port2Num := uint16(4332)
	port2Proto := EnclaveContainerPortProtocol_TCP
	expectedPort2DockerSpec := nat.Port(fmt.Sprintf(
		"%v/%v",
		port2Num,
		port2Proto,
	))
	expectedPort2EnclaveContainerPort, err := NewEnclaveContainerPort(port2Num, port2Proto)
	require.NoError(t, err, "Unexpected error occurred building the port 2 spec")

	privatePorts := map[string]*EnclaveContainerPort{
		port1Id: expectedPort1EnclaveContainerPort,
		port2Id: expectedPort2EnclaveContainerPort,
	}

	portIdsForDockerPortObjs, publishSpecs, err := getPortMapsBeforeContainerStart(privatePorts)
	require.NoError(t, err, "Unexpected error occurred getting the port maps before container start")
	require.Equal(t, len(privatePorts), len(portIdsForDockerPortObjs))
	require.Equal(t, len(privatePorts), len(publishSpecs))

	// Verify port 1 objs
	actualPort1Id, found := portIdsForDockerPortObjs[expectedPort1DockerSpec]
	require.True(t, found, "The result Docker port objs map didn't contain ID '%v'", port1Id)
	require.Equal(t, port1Id, actualPort1Id)
	_, found = publishSpecs[expectedPort1DockerSpec]
	require.True(t, found, "The result publish specs map didn't contain Docker port spec '%v'", expectedPort1DockerSpec)

	// Verify port 2 objs
	actualPort2Id, found := portIdsForDockerPortObjs[expectedPort2DockerSpec]
	require.True(t, found, "The result Docker port objs map didn't contain ID '%v'", port2Id)
	require.Equal(t, port2Id, actualPort2Id)
	_, found = publishSpecs[expectedPort2DockerSpec]
	require.True(t, found, "The result publish specs map didn't contain Docker port spec '%v'", expectedPort2DockerSpec)
}

func TestDuplicatedDockerPorts(t *testing.T) {
	port1Id := "port1"
	expectedPort1EnclaveContainerPort, err := NewEnclaveContainerPort(uint16(1234), EnclaveContainerPortProtocol_TCP)
	require.NoError(t, err, "Unexpected error occurred building the port 1 spec")

	port2Id := "port2"
	expectedPort2EnclaveContainerPort, err := NewEnclaveContainerPort(uint16(1234), EnclaveContainerPortProtocol_TCP)
	require.NoError(t, err, "Unexpected error occurred building the port 2 spec")

	privatePorts := map[string]*EnclaveContainerPort{
		port1Id: expectedPort1EnclaveContainerPort,
		port2Id: expectedPort2EnclaveContainerPort,
	}
	_, _, err = getPortMapsBeforeContainerStart(privatePorts)
	require.Error(t, err, "Expected an error when the same port is declared twice, but none happened")
}

// ====================================================================================================
//                                    Condense Network Info Tests
// ====================================================================================================
func TestNormalPublicNetworkInfoCondensing(t *testing.T) {
	expectedPublicIpAddrStr := "127.0.0.1"
	expectedPublicIpAddr := net.ParseIP(expectedPublicIpAddrStr)
	require.NotNil(t, expectedPublicIpAddr, "Parsing expected IP address string '%v' failed", expectedPublicIpAddrStr)

	port1Id := "rpc"
	port1DockerObj := nat.Port("1234/tcp")
	port1EnclaveContainerPort, err := NewEnclaveContainerPort(uint16(1234), EnclaveContainerPortProtocol_TCP)
	require.NoError(t, err, "An unexpected error occurred creating the port 1 enclave container port object")

	port2Id := "http"
	port2DockerObj := nat.Port("4333/tcp")
	port2EnclaveContainerPort, err := NewEnclaveContainerPort(uint16(4333), EnclaveContainerPortProtocol_TCP)
	require.NoError(t, err, "An unexpected error occurred creating the port 2 enclave container port object")

	hostMachinePortBindings := map[nat.Port]*nat.PortBinding{
		port1DockerObj: {
			HostIP:   expectedPublicIpAddrStr,
			HostPort: "62723",
		},
		port2DockerObj: {
			HostIP:   expectedPublicIpAddrStr,
			HostPort: "62724",
		},
		// Should get ignored, because this wasn't declared in the private ports
		nat.Port("9923/udp"): {
			HostIP:   expectedPublicIpAddrStr,
			HostPort: "62725",
		},
	}
	privatePorts := map[string]*EnclaveContainerPort{
		port1Id: port1EnclaveContainerPort,
		port2Id: port2EnclaveContainerPort,
	}
	portIdsForDockerPortObjs := map[nat.Port]string{
		port1DockerObj: port1Id,
		port2DockerObj: port2Id,
	}
	actualPublicIpAddr, actualPublicPorts, err := condensePublicNetworkInfoFromHostMachineBindings(
		hostMachinePortBindings,
		privatePorts,
		portIdsForDockerPortObjs,
	)
	require.NoError(t, err, "Unexpected error when condensing network info from host machine bindings")
	require.Equal(t, expectedPublicIpAddr, actualPublicIpAddr)
	require.Equal(t, len(privatePorts), len(actualPublicPorts))

	port1PublicPort, found := actualPublicPorts[port1Id]
	require.True(t, found, "Public ports didn't have a key for ID '%v'", port1Id)
	require.Equal(t, port1EnclaveContainerPort.GetProtocol(), port1PublicPort.GetProtocol())

	port2PublicPort, found := actualPublicPorts[port2Id]
	require.True(t, found, "Public ports didn't have a key for ID '%v'", port2Id)
	require.Equal(t, port2EnclaveContainerPort.GetProtocol(), port2PublicPort.GetProtocol())
}

func TestErrorOnNoHostMachineBindings(t *testing.T) {
	port1Id := "rpc"
	port1DockerObj := nat.Port("1234/tcp")
	port1EnclaveContainerPort, err := NewEnclaveContainerPort(uint16(1234), EnclaveContainerPortProtocol_TCP)
	require.NoError(t, err, "An unexpected error occurred creating the port 1 enclave container port object")

	hostMachinePortBindings := map[nat.Port]*nat.PortBinding{}
	privatePorts := map[string]*EnclaveContainerPort{
		port1Id: port1EnclaveContainerPort,
	}
	portIdsForDockerPortObjs := map[nat.Port]string{
		port1DockerObj: port1Id,
	}
	_, _, err = condensePublicNetworkInfoFromHostMachineBindings(
		hostMachinePortBindings,
		privatePorts,
		portIdsForDockerPortObjs,
	)
	require.Error(t, err, "Expected error when empty host machine port bindings map is supplied, but none was thrown")
}

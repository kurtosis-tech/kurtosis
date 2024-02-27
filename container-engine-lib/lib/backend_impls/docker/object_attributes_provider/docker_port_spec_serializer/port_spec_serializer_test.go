package docker_port_spec_serializer

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var portWaitForTest = port_spec.NewWait(5 * time.Second)

func TestValidSerDe(t *testing.T) {
	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	serialized, err := SerializePortSpecs(input)
	require.NoError(t, err, "An unexpected error occurred serializing the input")

	output, err := DeserializePortSpecs(serialized.GetString())
	require.NoError(t, err, "An unexpected error occurred deserializing the serialized input")
	require.Equal(t, len(input), len(output))

	for actualPortId, actualPortSpec := range output {
		expectedPortSpec, found := input[actualPortId]
		require.True(t, found, "Found port ID '%v' in the output that wasn't in the input", actualPortId)

		require.Equal(t, expectedPortSpec.GetNumber(), actualPortSpec.GetNumber(), "Actual port number for port '%v' doesn't match input", actualPortId)
		require.Equal(t, expectedPortSpec.GetTransportProtocol(), actualPortSpec.GetTransportProtocol(), "Actual port protocol for port '%v' doesn't match input", actualPortId)

		//TODO uncomment this when we include the wait serialization/deserialization
		/*require.Equal(t, expectedPortSpec.GetWait().GetEnable(), actualPortSpec.GetWait().GetEnable(), "Actual port wait enable field for port '%v' doesn't match input", actualPortId)
		require.Equal(t, expectedPortSpec.GetWait().GetInitialDelay().String(), actualPortSpec.GetWait().GetInitialDelay().String(), "Actual port wait initial delay field for port '%v' doesn't match input", actualPortId)
		require.Equal(t, expectedPortSpec.GetWait().GetTimeout().String(), actualPortSpec.GetWait().GetTimeout().String(), "Actual port wait initial delay field for port '%v' doesn't match input", actualPortId)*/
	}
}

// TODO REMOVE THIS AFTER JUNE 20, 2022 WHEN NOBODY IS USING OLD PORT SPECS
func TestDeserializeOldPortSpecs(t *testing.T) {
	eth2ContainerPortSpecStr := "rpc.8545-TCP_ws.8546-TCP_tcpDiscovery.30303-TCP_udpDiscovery.30303-UDP"
	_, err := DeserializePortSpecs(eth2ContainerPortSpecStr)
	require.NoError(t, err, "An unexpected error occurred deserializing the long-but-valid old port spec")
}

func TestValidLongDeserialization(t *testing.T) {
	eth2ContainerPortSpecStr := "rpc:8545/TCP,ws:8546/TCP,tcpDiscovery:30303/TCP,udpDiscovery:30303/UDP"
	_, err := DeserializePortSpecs(eth2ContainerPortSpecStr)
	require.NoError(t, err, "An unexpected error occurred deserializing the long-but-valid port spec")
}

func TestDisallowedCharsSerialization(t *testing.T) {
	for disallowedChar := range disallowedPortIdChars {
		portId := "ohyeah" + disallowedChar
		portNum := uint16(45)
		portProtocol := port_spec.TransportProtocol_TCP
		portSpec, err := port_spec.NewPortSpec(portNum, portProtocol, "", portWaitForTest, "")
		require.NoError(t, err, "An unexpected error occurred creating port spec for port with ID '%v'", portId)

		ports := map[string]*port_spec.PortSpec{
			portId: portSpec,
		}

		_, err = SerializePortSpecs(ports)
		require.Error(t, err, "Expected an error when serialized port with ID '%v' but none was thrown", portId)
	}
}

func TestDuplicatedPortNumDifferentProtoSerialization(t *testing.T) {
	dupedPortNum := uint16(77)

	port1Id := "port1"
	port1Num := dupedPortNum
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := dupedPortNum
	port2Protocol := port_spec.TransportProtocol_UDP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	_, err = SerializePortSpecs(input)
	require.NoError(t, err, "Expected two ports with the same number but different protocols to serialize successfully, but an error was thrown")
}

func TestDuplicatedPortNumSameProtoSerialization(t *testing.T) {
	dupedPortNum := uint16(77)

	port1Id := "port1"
	port1Num := dupedPortNum
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := dupedPortNum
	port2Protocol := port_spec.TransportProtocol_TCP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", portWaitForTest, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	_, err = SerializePortSpecs(input)
	require.Error(t, err, "Expected an error when serializing a ports map that has two ports with the same port num, but none was thrown")
}

func TestBlankPortDeserialization(t *testing.T) {
	input := "my_port:23/TCP,"
	_, err := DeserializePortSpecs(input)
	require.Error(t, err, "Expected an error when deserializing a string with an empty entry, but none was report")
}

func TestBlankPortSpecDeserialization(t *testing.T) {
	input := "my_port:23/TCP,your_port:"
	_, err := DeserializePortSpecs(input)
	require.Error(t, err, "Expected an error when deserializing a string with a port without a spec, but none was report")
}

func TestMissingPortNumDeserialization(t *testing.T) {
	input := "my_port:23/TCP,your_port:/UDP"
	_, err := DeserializePortSpecs(input)
	require.Error(t, err, "Expected an error when deserializing a string with a port without a number, but none was report")
}

func TestNonnumericPortNumDeserialization(t *testing.T) {
	input := "my_port:23/TCP,your_port:abcd/UDP"
	_, err := DeserializePortSpecs(input)
	require.Error(t, err, "Expected an error when deserializing a string with a port with a non-numeric portnum, but none was report")
}

func TestInvalidProtocolDeserialization(t *testing.T) {
	input := "my_port:23/TCP,your_port:27/nonexistent"
	_, err := DeserializePortSpecs(input)
	require.Error(t, err, "Expected an error when deserializing a string with a port with an invalid protocol, but none was report")
}

func TestDeserializePortSpecStrUsingDelimiters_PortSpecWithIncorrectNumberOfFragments(t *testing.T) {
	invalidInputWithOneFragments := "myport:23"
	invalidInputWithFourFragments := "myport:23/tcp/https/abc"

	_, err := deserializePortSpecStrUsingDelimiters(
		invalidInputWithOneFragments,
		portSpecsSeparator,
		portIdAndInfoSeparator,
		portNumAndProtocolSeparator,
	)
	require.NotNil(t, err, fmt.Sprintf("Expected error to be not nil for %v", invalidInputWithOneFragments))
	require.ErrorContains(t, err, "Expected splitting port spec string '23' to yield '2' to '3' fragments but got '1'")

	_, err = deserializePortSpecStrUsingDelimiters(
		invalidInputWithFourFragments,
		portSpecsSeparator,
		portIdAndInfoSeparator,
		portNumAndProtocolSeparator,
	)
	require.NotNil(t, err, fmt.Sprintf("Expected error to be not nil for %v", invalidInputWithFourFragments))
	require.ErrorContains(t, err, "Expected splitting port spec string '23/tcp/https/abc' to yield '2' to '3' fragments but got '4'")
}

func TestNoPortProtosHaveDisallowedChars(t *testing.T) {
	for _, portProtoStr := range port_spec.TransportProtocolStrings() {
		for illegalTransportProtocolStr := range disallowedPortIdChars {
			require.False(t, strings.Contains(portProtoStr, illegalTransportProtocolStr))
		}
	}
}

func TestValidatePortSpec_ValidApplicationProtocol(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_UDP, "H-ttp.2", portWaitForTest, "")
	err := validatePortSpec("PortId", spec)
	require.Nil(t, err, "Error cannot be nil")
}

func TestValidatePortSpec_InvalidPortId(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_TCP, "https", portWaitForTest, "")
	err := validatePortSpec(",portid/", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, fmt.Sprintf("Port ID '%v' contains disallowed char '%v'", ",portid/", portSpecsSeparator))
}

func TestValidatePortSpec_InvalidApplicationProtocol(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_TCP, "/https,", portWaitForTest, "")
	err := validatePortSpec("PortId", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, fmt.Sprintf("Application Protocol '%v' associated with port ID '%v' contains disallowed char '%v'", "/https,", "PortId", portNumAndProtocolSeparator))

	spec, _ = port_spec.NewPortSpec(100, port_spec.TransportProtocol_UDP, " H-ttp.2", portWaitForTest, "")
	err = validatePortSpec("PortId", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, "application protocol ' H-ttp.2' associated with port ID 'PortId' contains invalid character(s). It must only contain [a-zA-Z0-9+.-]")
}

func TestSerializeMethod_ValidPortSpecs(t *testing.T) {
	specs := map[string]*port_spec.PortSpec{}
	portOne, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_TCP, "", portWaitForTest, "")
	portTwo, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_UDP, "https", portWaitForTest, "")

	specs["portOne"] = portOne
	specs["portTwo"] = portTwo

	possibleValidExpectedSerializedPortSpec := []string{
		"portOne:3333/TCP,portTwo:3333/UDP/https", "portTwo:3333/UDP/https,portOne:3333/TCP",
	}

	actualLabelValue, err := SerializePortSpecs(specs)
	require.Nil(t, err)
	// because map is not ordered therefore need to test against two possible valid outcomes
	require.Contains(t, possibleValidExpectedSerializedPortSpec, actualLabelValue.GetString())
}

func TestDeSerializeMethod_ValidPortSpecs(t *testing.T) {
	expectedSpecs := map[string]*port_spec.PortSpec{}

	//TODO use portWaitForTest when we include the wait serialization/deserialization
	expectedPortOne, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_TCP, "", nil, "")
	expectedPortTwo, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_UDP, "https", nil, "")
	expectedSpecs["portOne"] = expectedPortOne
	expectedSpecs["portTwo"] = expectedPortTwo

	portSpecStr1 := "portOne:3333/TCP,portTwo:3333/UDP/https"
	actualPortSpec1, err := DeserializePortSpecs(portSpecStr1)
	require.Nil(t, err)
	require.Equal(t, expectedSpecs, actualPortSpec1)

	portSpecStr2 := "portOne:3333/TCP,portTwo:3333/UDP/https"
	actualPortSpec2, err := DeserializePortSpecs(portSpecStr2)
	require.Nil(t, err)
	require.Equal(t, expectedSpecs, actualPortSpec2)
}

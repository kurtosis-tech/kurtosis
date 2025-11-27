package kubernetes_port_spec_serializer

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestValidSerDe(t *testing.T) {
	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", nil, "")
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
	}
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
		portSpec, err := port_spec.NewPortSpec(portNum, portProtocol, "", nil, "")
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
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := dupedPortNum
	port2Protocol := port_spec.TransportProtocol_UDP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	_, err = SerializePortSpecs(input)
	require.NoError(t, err, "Expected two ports with the same number but different protocols to serialize successfully, but an error was thrown")
}

func TestWaitSerDer(t *testing.T) {

	port0Id := "port0"
	port0Num := uint16(22)
	port0Protocol := port_spec.TransportProtocol_TCP
	port0Spec, err := port_spec.NewPortSpec(port0Num, port0Protocol, "", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 0 spec")

	port1Id := "port1"
	port1Num := uint16(123)
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", port_spec.NewWait(6*time.Second), "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(321)
	port2Protocol := port_spec.TransportProtocol_UDP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "http", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	port3Id := "port3"
	port3Num := uint16(11)
	port3Protocol := port_spec.TransportProtocol_TCP
	port3Spec, err := port_spec.NewPortSpec(port3Num, port3Protocol, "postgres", port_spec.NewWait(5*time.Millisecond), "")
	require.NoError(t, err, "An unexpected error occurred creating port 3 spec")

	portIds := []string{port0Id, port1Id, port2Id, port3Id}
	input := map[string]*port_spec.PortSpec{
		port0Id: port0Spec,
		port1Id: port1Spec,
		port2Id: port2Spec,
		port3Id: port3Spec,
	}

	serializedPortSpecs, err := SerializePortSpecs(input)
	require.NoError(t, err)
	derializedPortSpecs, err := DeserializePortSpecs(serializedPortSpecs.GetString())
	require.NoError(t, err)
	for _, portId := range portIds {
		require.Equal(t, derializedPortSpecs[portId], input[portId])
	}
}

func TestDuplicatedPortNumSameProtoSerialization(t *testing.T) {
	dupedPortNum := uint16(77)

	port1Id := "port1"
	port1Num := dupedPortNum
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", nil, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := dupedPortNum
	port2Protocol := port_spec.TransportProtocol_TCP
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, "", nil, "")
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

func TestNoPortProtosHaveDisallowedChars(t *testing.T) {
	for _, portProtoStr := range port_spec.TransportProtocolStrings() {
		for illegalTransportProtocolStr := range disallowedPortIdChars {
			require.False(t, strings.Contains(portProtoStr, illegalTransportProtocolStr))
		}
	}
}

func TestValidatePortSpec_ValidApplicationProtocol(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_UDP, "H-ttp.2", nil, "")
	err := validatePortSpec("PortId", spec)
	require.Nil(t, err, "Error cannot be nil")
}

func TestValidatePortSpec_InvalidPortId(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_TCP, "https", nil, "")
	err := validatePortSpec(",portid/", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, fmt.Sprintf("Port ID '%v' contains disallowed char '%v'", ",portid/", portSpecsSeparator))
}

func TestValidatePortSpec_InvalidApplicationProtocol(t *testing.T) {
	spec, _ := port_spec.NewPortSpec(100, port_spec.TransportProtocol_TCP, "/https,", nil, "")
	err := validatePortSpec("PortId", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, fmt.Sprintf("Application Protocol '%v' associated with port ID '%v' contains disallowed char '%v'", "/https,", "PortId", portNumAndProtocolSeparator))

	spec, _ = port_spec.NewPortSpec(100, port_spec.TransportProtocol_UDP, " H-ttp.2", nil, "")
	err = validatePortSpec("PortId", spec)
	require.NotNil(t, err, "Error cannot be nil")
	require.ErrorContains(t, err, "application protocol ' H-ttp.2' associated with port ID 'PortId' contains invalid character(s). It must only contain [a-zA-Z0-9+.-]")
}

func TestSerializeMethod_ValidPortSpecs(t *testing.T) {
	specs := map[string]*port_spec.PortSpec{}
	portOne, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_TCP, "", nil, "")
	portTwo, _ := port_spec.NewPortSpec(3333, port_spec.TransportProtocol_UDP, "https", nil, "")

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

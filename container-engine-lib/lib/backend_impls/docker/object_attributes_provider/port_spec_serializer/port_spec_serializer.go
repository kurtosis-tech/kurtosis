package port_spec_serializer

import (
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"strconv"
	"strings"
)

const (
	portIdAndInfoSeparator      = ":"
	portNumAndProtocolSeparator = "/"
	portSpecsSeparator          = ","

	// TODO DELETE TEHSE AFTER JUNE 20, 2022 WHEN WE'RE CONFIDENT NOBODY'S USING THE OLD PORT SPECS!
	oldPortIdAndInfoSeparator      = "."
	oldPortNumAndProtocolSeparator = "-"
	oldPortSpecsSeparator          = "_"

	expectedNumPortIdAndSpecFragments      = 2
	expectedNumPortNumAndProtocolFragments = 2
	portUintBase                           = 10
	portUintBits                           = 16

	// The maximum number of bytes that a label value can be
	// See https://github.com/docker/for-mac/issues/2208
	maxLabelValueBytes = 65518
)

// "Set" of the disallowed characters for a port ID
var disallowedPortIdChars = map[string]bool{
	portIdAndInfoSeparator:      true,
	portNumAndProtocolSeparator: true,
	portSpecsSeparator:          true,
}

// NOTE: We use a custom serialization format here (rather than, e.g., JSON) because there's a max label value size
//  so brevity is important here
func SerializePortSpecs(ports map[string]*port_spec.PortSpec) (*docker_label_value.DockerLabelValue, error) {
	portIdAndSpecStrs := []string{}
	usedPortSpecStrs := map[string]string{}
	for portId, portSpec := range ports {
		for disallowedChar := range disallowedPortIdChars {
			if strings.Contains(portId, disallowedChar) {
				return nil, stacktrace.NewError("Port ID '%v' contains disallowed char '%v'", portId, disallowedChar)
			}
			protocolStr := portSpec.GetProtocol().String()
			if strings.Contains(protocolStr, disallowedChar) {
				return nil, stacktrace.NewError(
					"Port spec protocol '%v' for port with ID '%v' contains disallowed char '%v'",
					protocolStr,
					portId,
					disallowedChar,
				)
			}
		}
		portNum := portSpec.GetNumber()
		portProtocol := portSpec.GetProtocol()
		if !portProtocol.IsAPortProtocol() {
			return nil, stacktrace.NewError("Unrecognized port protocol '%v'", portProtocol.String())
		}
		portSpecStr := fmt.Sprintf(
			"%v%v%v",
			portNum,
			portNumAndProtocolSeparator,
			portProtocol.String(),
		)

		if previousPortId, found := usedPortSpecStrs[portSpecStr]; found {
			return nil, stacktrace.NewError(
				"Port '%v' declares spec string '%v', but that spec string is already in use for port '%v'",
				portId,
				portSpecStr,
				previousPortId,
			)
		}
		usedPortSpecStrs[portSpecStr] = portId

		portIdAndSpecStr := fmt.Sprintf(
			"%v%v%v",
			portId,
			portIdAndInfoSeparator,
			portSpecStr,
		)
		portIdAndSpecStrs = append(portIdAndSpecStrs, portIdAndSpecStr)
	}
	resultStr := strings.Join(portIdAndSpecStrs, portSpecsSeparator)
	numResultBytes := len([]byte(resultStr))
	if numResultBytes > maxLabelValueBytes {
		return nil, stacktrace.NewError(
			"The port specs label value string is %v bytes long, but the max number of label value bytes is %v; the number of ports this container is listening on must be reduced",
			numResultBytes,
			maxLabelValueBytes,
		)
	}
	result, err := docker_label_value.CreateNewDockerLabelValue(resultStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Docker label value from string '%v'", resultStr)
	}
	return result, nil
}

func DeserializePortSpecs(specsStr string) (map[string]*port_spec.PortSpec, error) {
	resultUsingNewDelimiters, err := deserializePortSpecStrUsingDelimiters(
		specsStr,
		portSpecsSeparator,
		portIdAndInfoSeparator,
		portNumAndProtocolSeparator,
	)
	if err == nil {
		return resultUsingNewDelimiters, nil
	}

	// TODO DELETE THIS CHECK AFTER JUNE 20, 2022 WHEN WE'RE CONFIDENT NOBODY WILL HAVE THE OLD PORT SPEC!!!
	resultUsingOldDelimiters, err := deserializePortSpecStrUsingDelimiters(
		specsStr,
		oldPortSpecsSeparator,
		oldPortIdAndInfoSeparator,
		oldPortNumAndProtocolSeparator,
	)
	if err == nil {
		return resultUsingOldDelimiters, nil
	}

	return nil, stacktrace.Propagate(
		err,
		"Couldn't deserialize old port spec string '%v' using old port spec delimiters",
		specsStr,
	)
}

func deserializePortSpecStrUsingDelimiters(
	specsStr string,
	portSpecsSeparator string,
	portIdAndInfoSeparator string,
	portNumAndProtocolSeparator string,
) (
	map[string]*port_spec.PortSpec,
	error,
) {
	result := map[string]*port_spec.PortSpec{}
	if specsStr == "" {
		return result, nil
	}

	portIdAndSpecStrs := strings.Split(specsStr, portSpecsSeparator)
	for _, portIdAndSpecStr := range portIdAndSpecStrs {
		portIdAndSpecFragments := strings.Split(portIdAndSpecStr, portIdAndInfoSeparator)
		numPortIdAndSpecFragments := len(portIdAndSpecFragments)
		if numPortIdAndSpecFragments != expectedNumPortIdAndSpecFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port ID & spec string '%v' to yield %v fragments but got %v",
				portIdAndSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portId := portIdAndSpecFragments[0]
		portSpecStr := portIdAndSpecFragments[1]

		portNumAndProtocolFragments := strings.Split(portSpecStr, portNumAndProtocolSeparator)
		numPortNumAndProtocolFragments := len(portNumAndProtocolFragments)
		if numPortNumAndProtocolFragments != expectedNumPortNumAndProtocolFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port num & protocol string '%v' to yield %v fragments but got %v",
				portSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portNumStr := portNumAndProtocolFragments[0]
		portProtocolStr := portNumAndProtocolFragments[1]

		portNumUint64, err := strconv.ParseUint(portNumStr, portUintBase, portUintBits)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred parsing port num string '%v' to uint with base %v and %v bits",
				portNumStr,
				portUintBase,
				portUintBits,
			)
		}
		portNumUint16 := uint16(portNumUint64)
		portProtocol, err := port_spec.PortProtocolString(portProtocolStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting port protocol string '%v' to a port protocol enum", portProtocolStr)
		}

		portSpec, err := port_spec.NewPortSpec(portNumUint16, portProtocol)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating port spec object from ID & spec string '%v'",
				portIdAndSpecStr,
			)
		}

		result[portId] = portSpec
	}
	return result, nil
}
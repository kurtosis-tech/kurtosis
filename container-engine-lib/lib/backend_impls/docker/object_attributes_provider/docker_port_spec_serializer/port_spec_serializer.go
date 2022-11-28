package docker_port_spec_serializer

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
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

	maybeApplicationProtocolFragment  = 3
	expectedNumPortIdAndSpecFragments = 2
	minExpectedNumPortNumAndProtocol  = 2
	maxExpectedNumPortNumAndProtocol  = 3
	portUintBase                      = 10
	portUintBits                      = 16

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

/*
  This method is used to validate port id - it must not have any disallowed characters.
  This is not needed for protocol, and application protocol because they are defined as enums.
*/
func validatePortSpec(portId string) error {
	validator := regexp.MustCompile(fmt.Sprintf("[%v%v%v]", portNumAndProtocolSeparator, portIdAndInfoSeparator, portSpecsSeparator))

	// validate portId
	shouldBeEmptyForPortId := validator.FindString(portId)

	if len(shouldBeEmptyForPortId) > 0 {
		return stacktrace.NewError(
			"Port ID '%v' contains disallowed char '%v'",
			portId,
			shouldBeEmptyForPortId,
		)
	}

	return nil
}

// NOTE: We use a custom serialization format here (rather than, e.g., JSON) because there's a max label value size
//  so brevity is important here
func SerializePortSpecs(ports map[string]*port_spec.PortSpec) (*docker_label_value.DockerLabelValue, error) {
	portIdAndSpecStrs := []string{}
	usedPortSpecStrs := map[string]string{}

	for portId, portSpec := range ports {
		err := validatePortSpec(portId)

		if err != nil {
			return nil, err
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

		// add application protocol to the label value if present
		if portSpec.GetApplicationProtocol() != nil {
			portIdAndSpecStr = fmt.Sprintf("%v/%v", portIdAndSpecStr, portSpec.GetApplicationProtocol().String())
		}
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
		"Failed to deserialize port spec string '%v' after trying both current and old port spec delimiters",
		specsStr,
	)
}

func createPortSpec(
	number uint16,
	protocol port_spec.PortProtocol,
	applicationProtocol string,
) (*port_spec.PortSpec, error) {
	if applicationProtocol != "" {
		appProtocol, _ := port_spec.ApplicationProtocolString(applicationProtocol)
		return port_spec.NewPortSpec(number, protocol, appProtocol)
	}

	return port_spec.NewPortSpec(number, protocol)
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
		if numPortNumAndProtocolFragments < minExpectedNumPortNumAndProtocol || numPortNumAndProtocolFragments > maxExpectedNumPortNumAndProtocol {
			return nil, stacktrace.NewError(
				"Expected splitting port num & protocol string '%v' to yield '%v' or '%v' fragments but got %v",
				portSpecStr,
				minExpectedNumPortNumAndProtocol,
				maxExpectedNumPortNumAndProtocol,
				numPortNumAndProtocolFragments,
			)
		}

		portApplicationProtocolStr := ""
		portNumStr := portNumAndProtocolFragments[0]
		portProtocolStr := portNumAndProtocolFragments[1]

		if numPortNumAndProtocolFragments == maybeApplicationProtocolFragment {
			portApplicationProtocolStr = portNumAndProtocolFragments[2]
		}

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

		portSpec, err := createPortSpec(portNumUint16, portProtocol, portApplicationProtocolStr)

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

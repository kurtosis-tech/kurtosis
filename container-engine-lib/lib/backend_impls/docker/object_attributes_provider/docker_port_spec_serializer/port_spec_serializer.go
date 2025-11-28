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

	numPortSpecFragmentsWithApplicationProtocol = 3
	expectedNumPortIdAndSpecFragments           = 2
	minExpectedPortSpecFragments                = 2
	maxExpectedPortSpecFragments                = 3
	portUintBase                                = 10
	portUintBits                                = 16

	portNumIndex                 = 0
	portProtocolIndex            = 1
	portApplicationProtocolIndex = 2
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

var disallowedCharactersMatcher = regexp.MustCompile(fmt.Sprintf("[%v%v%v]", portNumAndProtocolSeparator, portIdAndInfoSeparator, portSpecsSeparator))
var validApplicationProtocolMatcher = regexp.MustCompile(`^[a-zA-Z0-9+.-]*$`)

// NOTE: We use a custom serialization format here (rather than, e.g., JSON) because there's a max label value size
// so brevity is important here
func SerializePortSpecs(ports map[string]*port_spec.PortSpec) (*docker_label_value.DockerLabelValue, error) {
	portIdAndSpecStrs := []string{}
	usedPortSpecStrs := map[string]string{}

	for portId, portSpec := range ports {
		err := validatePortSpec(portId, portSpec)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error occurred while validating port spec '%+v'", portSpec)
		}

		portNum := portSpec.GetNumber()
		portProtocol := portSpec.GetTransportProtocol()
		if !portProtocol.IsATransportProtocol() {
			return nil, stacktrace.NewError("Unrecognized port protocol '%v'", portProtocol.String())
		}
		portSpecStr := fmt.Sprintf(
			"%v%v%v",
			portNum,
			portNumAndProtocolSeparator,
			portProtocol.String(),
		)

		// add application protocol to the label value if present
		maybeApplicationProtocol := portSpec.GetMaybeApplicationProtocol()
		if maybeApplicationProtocol != nil {
			portSpecStr = fmt.Sprintf("%v%v%v", portSpecStr, portNumAndProtocolSeparator, *maybeApplicationProtocol)
		}

		//TODO we are not serializing the port wait so far, because this in design stage and the name
		//TODO and even the fields could change, so we don't want to brake anything until we add the port wait configs to users
		//TODO we aren't serializing URL for now
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

	//TODO we are not deserializing the port wait so far, because this in design stage and the name
	//TODO and even the fields could change, so we don't want to brake anything until we add the port wait configs to users

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
		portSpecFragments := strings.Split(portSpecStr, portNumAndProtocolSeparator)
		numPortSpecFragments := len(portSpecFragments)
		if numPortSpecFragments < minExpectedPortSpecFragments || numPortSpecFragments > maxExpectedPortSpecFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port spec string '%v' to yield '%v' to '%v' fragments but got '%v'",
				portSpecStr,
				minExpectedPortSpecFragments,
				maxExpectedPortSpecFragments,
				numPortSpecFragments,
			)
		}

		portNumStr := portSpecFragments[portNumIndex]
		portProtocolStr := portSpecFragments[portProtocolIndex]

		portApplicationProtocolStr := ""
		if numPortSpecFragments == numPortSpecFragmentsWithApplicationProtocol {
			portApplicationProtocolStr = portSpecFragments[portApplicationProtocolIndex]
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
		portProtocol, err := port_spec.TransportProtocolString(portProtocolStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting port protocol string '%v' to a port protocol enum", portProtocolStr)
		}

		//TODO we are passing nil wait so far because the wait's serialization/deserialization logic is not added yet because
		//TODO the port wait feature is in the design stage and its name and fields could change,
		//TODO we will include this in a next PR
		portSpec, err := port_spec.NewPortSpec(portNumUint16, portProtocol, portApplicationProtocolStr, nil, "")
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

// validatePortSpec is used to validate port id - it must not have any disallowed characters.
// This is not needed for protocol, because it is defined as enums.
func validatePortSpec(portId string, spec *port_spec.PortSpec) error {
	// validate port id - it should not contain disallowed characters
	firstDisallowedCharacterInPortId := disallowedCharactersMatcher.FindString(portId)
	if len(firstDisallowedCharacterInPortId) > 0 {
		return stacktrace.NewError(
			"Port ID '%v' contains disallowed char '%v'",
			portId,
			firstDisallowedCharacterInPortId,
		)
	}

	// validate application protocol - it should not contain disallowed characters
	maybeApplicationProtocol := spec.GetMaybeApplicationProtocol()
	if maybeApplicationProtocol != nil {
		firstDisallowedCharacterInApplicationProtocol := disallowedCharactersMatcher.FindString(*maybeApplicationProtocol)
		if len(firstDisallowedCharacterInApplicationProtocol) > 0 {
			return stacktrace.NewError(
				"Application Protocol '%v' associated with port ID '%v' contains disallowed char '%v'",
				*maybeApplicationProtocol,
				portId,
				firstDisallowedCharacterInApplicationProtocol,
			)
		}

		doesApplicationProtocolContainsValidChar := validApplicationProtocolMatcher.MatchString(*maybeApplicationProtocol)
		if !doesApplicationProtocolContainsValidChar {
			return stacktrace.NewError(
				"application protocol '%v' associated with port ID '%v' contains invalid character(s). It must only contain [a-zA-Z0-9+.-]",
				*maybeApplicationProtocol,
				portId)
		}
	}

	return nil
}

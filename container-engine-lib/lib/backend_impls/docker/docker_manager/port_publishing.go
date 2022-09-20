package docker_manager

// ====================================================================================================
//                                            Interface
// ====================================================================================================
// "Enum" dictating the various types of port publishing available; this enum will be used for downcasting the
//  PortPublishSpec
type portPublishSpecType string
const (
	// The port should not be published to the host machine at all
	noPublishing portPublishSpecType = "NONE"

	// The port should be published to the host machine on a port that is chosen automatically by the container engine
	automaticPublishing portPublishSpecType = "AUTOMATIC"

	// The port should be published to a specified, known port on the host machine
	manualPublishing portPublishSpecType = "MANUAL"
)

// To get an instance of this type, use the NewNoPublishingSpec, NewAutomaticPublishingSpec, etc. functions
type PortPublishSpec interface {
	// Used internally, for downcasting the interface to the appropriate implementation
	getType() portPublishSpecType

	// If true, after the container starts then this port must be found in the host machine ports list and reported
	//  to the user as part of container start
	mustBeFoundAfterContainerStart() bool
}

// ====================================================================================================
//                                         Simple Publish Spec
// ====================================================================================================
// A PortPublishSpec implementation that only contains a type
type simplePortPublishSpec struct {
	publishType                   portPublishSpecType
	shouldFindAfterContainerStart bool
}
func (spec *simplePortPublishSpec) getType() portPublishSpecType {
	return spec.publishType
}

func (spec *simplePortPublishSpec) mustBeFoundAfterContainerStart() bool {
	return spec.shouldFindAfterContainerStart
}

// Returns a PortPublishSpec indicating that the port shouldn't be published to the host machine at all
func NewNoPublishingSpec() PortPublishSpec {
	return &simplePortPublishSpec{
		publishType:                   noPublishing,
		shouldFindAfterContainerStart: false,
	}
}

// Returns a PortPublishSpec indicating that the port should be published to an automatically-assigned port on the host machine
func NewAutomaticPublishingSpec() PortPublishSpec {
	return &simplePortPublishSpec{
		publishType:                   automaticPublishing,
		shouldFindAfterContainerStart: true,
	}
}

// ====================================================================================================
//                                         Manual Publish Spec
// ====================================================================================================
// A PortPublishSpec implementation, used for the manualPublishing option type, that also contains the manual port to publish to
type manuallySpecifiedPortPublishSpec struct {
	simplePortPublishSpec

	hostMachinePortNum uint16
}
func (option *manuallySpecifiedPortPublishSpec) getHostMachinePortNum() uint16 {
	return option.hostMachinePortNum
}

// Returns a PortPublishSpec indicating that the port should be published to the given port on the host machine
func NewManualPublishingSpec(hostMachinePortNum uint16) PortPublishSpec {
	return &manuallySpecifiedPortPublishSpec{
		simplePortPublishSpec: simplePortPublishSpec{
			publishType:                   manualPublishing,
			shouldFindAfterContainerStart: true,
		},
		hostMachinePortNum: hostMachinePortNum,
	}
}

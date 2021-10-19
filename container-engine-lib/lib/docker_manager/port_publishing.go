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

	// The port should be published to an ephemeral port on the host machine
	automaticPublishing portPublishSpecType = "AUTOMATIC"

	// The port should be published to the manually-specified port on the host machine
	manualPublishing portPublishSpecType = "MANUAL"
)

// To get an instance of this type, use the NewNoPublishingSpec, NewAutomaticPublishingSpec, etc. functions
type PortPublishSpec interface {
	// Used internally, for downcasting the interface to the appropriate implementation
	getType() portPublishSpecType
}

// ====================================================================================================
//                                         Simple Publish Spec
// ====================================================================================================
// A PortPublishSpec implementation that only contains a type
type simplePortPublishSpec struct {
	publishType portPublishSpecType
}
func (spec *simplePortPublishSpec) getType() portPublishSpecType {
	return spec.publishType
}

// Returns a PortPublishSpec indicating that the port shouldn't be published to the host machine at all
func NewNoPublishingSpec() PortPublishSpec {
	return &simplePortPublishSpec{publishType: noPublishing}
}

// Returns a PortPublishSpec indicating that the port should be published to an automatically-assigned port on the host machine
func NewAutomaticPublishingSpec() PortPublishSpec {
	return &simplePortPublishSpec{publishType: automaticPublishing}
}

// ====================================================================================================
//                                         Manual Publish Spec
// ====================================================================================================
// A PortPublishSpec implementation, used for the manualPublishing option type, that also contains the manual port to publish to
type manualPublishingOption struct {
	simplePortPublishSpec

	hostMachinePortSpec string
}
func (option *manualPublishingOption) getHostMachinePortSpec() string {
	return option.hostMachinePortSpec
}

// Returns a PortPublishSpec indicating that the port should be published to the given port on the host machine
func NewManualPublishingSpec(hostMachinePort string) PortPublishSpec {
	return &manualPublishingOption{
		simplePortPublishSpec: simplePortPublishSpec{
			publishType: manualPublishing,
		},
		hostMachinePortSpec:        hostMachinePort,
	}
}

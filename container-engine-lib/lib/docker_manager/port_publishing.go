package docker_manager

// "Enum" dictating the various types of port publishing available; this enum will be used for downcasting the
//  PortPublishingOption
type portPublishingOptionType string
const (
	// The port should not be published to the host machine at all
	noPublishing portPublishingOptionType = "NONE"

	// The port should be publlished to an ephemeral port on the host machine
	automaticPublishing portPublishingOptionType = "AUTOMATIC"

	// The port should be published to the manually-specified port on the host machine
	manualPublishing portPublishingOptionType = "MANUAL"
)

// To get an instance of this type, use the
type PortPublishingOption interface {
	// Used internally, for downcasting the interface to the appropriate implementation
	getType() portPublishingOptionType
}

// A PortPublishingOption implementation that only contains a type
type simplePortPublishingOption struct {
	publishType portPublishingOptionType
}
func (option *simplePortPublishingOption) getType() portPublishingOptionType {
	return option.publishType
}

// Returns a PortPublishingOption indicating that the port shouldn't be published to the host machine at all
func NewNoPortPublishingOption() PortPublishingOption {
	return &simplePortPublishingOption{publishType: noPublishing}
}

// Returns a PortPublishingOption indicating that the port should be published to an automatically-assigned port on the host machine
func NewAutomaticPortPublishingOption() PortPublishingOption {
	return &simplePortPublishingOption{publishType: automaticPublishing}
}

// A PortPublishingOption implementation, used for the manualPublishing option type, that also contains the manual port to publish to
type manualPublishingOption struct {
	simplePortPublishingOption

	hostMachinePortSpec string
}
func (option *manualPublishingOption) getHostMachinePortSpec() string {
	return option.hostMachinePortSpec
}

func NewManualPortPublishingOption(hostMachinePortSpec string) PortPublishingOption {
	return &manualPublishingOption{
		simplePortPublishingOption: simplePortPublishingOption{
			publishType: manualPublishing,
		},
		hostMachinePortSpec:        "",
	}
}


package services

// TODO Rename this ServiceInitializerCore
// Contains configuration determining what type of objects the ServiceFactory will produce
// This is implicitly a DockerContainerServiceFactoryConfig; we could abstract it easily if we wanted other foundations for services
type ServiceFactoryConfig interface {
	GetDockerImage() string

	GetUsedPorts() map[int]bool

	// TODO when Go gets generics, make the type of 'dependencies' be the same as the output of GetStartCommand
	// If Go had generics, dependencies should be of type []T
	GetStartCommand(publicIpAddr string, dependencies []Service) []string

	// If Go had generics, the return type would be T
	GetServiceFromIp(ipAddr string) Service

	// ========================================================================================
	// NOTE: These functions below for checking liveness will probably need to be moved to a separate
	//  LivenessChecker object or to the Service itself; they live right here for now because it makes implementing the
	//  API very easy - the user just defines one ServiceFactoryConfig implementation. The impetus for moving
	//  this would be when topology needs to be dynamic at testing time (i.e. a test can start and stop nodes)

	// TODO When Go gets generics, make the type of these args parameterized
	IsServiceUp(toCheck Service, dependencies []Service) bool

	// How long to wait for the service to start up before giving up
	GetStartupTimeoutMillis() int64
}


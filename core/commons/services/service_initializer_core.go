package services

// TODO When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
// Contains configuration determining what type of objects the ServiceInitializer will produce
// This is implicitly a DockerContainerServiceFactoryConfig; we could abstract it easily if we wanted other foundations for services
type ServiceInitializerCore interface {
	GetUsedPorts() map[int]bool

	// TODO when Go gets generics, make the type of 'dependencies' to be []N
	// If Go had generics, dependencies should be of type []T
	GetStartCommand(publicIpAddr string, dependencies []Service) []string

	// TODO When Go has generics, make this return type to be S
	GetServiceFromIp(ipAddr string) Service
}


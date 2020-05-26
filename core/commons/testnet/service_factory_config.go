package testnet

// Contains configuration determining what type of objects the ServiceFactory will produce
// This is implicitly a DockerContainerServiceFactoryConfig; we could abstract it easily if we wanted other foundations for services
type ServiceFactoryConfig interface {
	GetDockerImage() string

	GetUsedPorts() map[int]bool

	// TODO somehow, some way verify that the types of these two functions are equal
	// TODO This ipAddrOffset is a nasty hack that will go away when public-ips is gone!
	// If Go had generics, dependencies should be of type []T
	GetStartCommand(ipAddrOffset int, dependencies []Service) []string

	// If Go had generics, the return type would be T
	GetServiceFromIp(ipAddr string) Service
}


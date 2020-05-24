package commons

type ServiceNetwork struct {
	NetworkId string

	// If Go had generics, we'd make this object genericized and use that as the return type here
	Services map[int]Service

	ContainerIds map[int]string
}

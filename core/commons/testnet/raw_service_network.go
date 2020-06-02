package testnet

type RawServiceNetwork struct {
	// If Go had generics, we'd make this object genericized and use that as the return type here
	ServiceIPs map[int]string

	ContainerIds map[int]string
}

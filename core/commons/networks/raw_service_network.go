package networks

type RawServiceNetwork struct {
	ServiceIPs map[int]string

	ContainerIds map[int]string
}

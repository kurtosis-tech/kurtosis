package port_spec

//go:generate go run github.com/dmarkham/enumer -transform=snake-upper -trimprefix=PortProtocol_ -type=PortProtocol
type PortProtocol int
const (
	PortProtocol_TCP PortProtocol = iota
	PortProtocol_SCTP
	PortProtocol_UDP
)

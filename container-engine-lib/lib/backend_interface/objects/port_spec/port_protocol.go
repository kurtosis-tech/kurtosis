package port_spec

//go:generate go run github.com/dmarkham/enumer -transform=snake-upper -trimprefix=TransportProtocol_ -type=TransportProtocol
type TransportProtocol int

const (
	TransportProtocol_TCP TransportProtocol = iota
	TransportProtocol_SCTP
	TransportProtocol_UDP
)

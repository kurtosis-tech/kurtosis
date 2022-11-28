package port_spec

//go:generate go run github.com/dmarkham/enumer -transform=lower -type=ApplicationProtocol
type ApplicationProtocol int

const (
	HTTP ApplicationProtocol = iota
	HTTPS
)

package run

//go:generate go run github.com/dmarkham/enumer -type=Verbosity -transform=snake-upper
type Verbosity int

const (
	Brief Verbosity = iota
	Detailed
	Executable
)

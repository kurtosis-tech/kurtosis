package run

//go:generate go run github.com/dmarkham/enumer -type=Verbosity -transform=snake-upper
type ImageDownload int

const (
	never ImageDownload = iota
	always
	missing
)

package run

//go:generate go run github.com/dmarkham/enumer -type=Verbosity -transform=snake-upper
type ImageDownload int

const (
	always ImageDownload = iota
	missing
)

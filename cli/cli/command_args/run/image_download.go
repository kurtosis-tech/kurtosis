package run

//go:generate go run github.com/dmarkham/enumer -type=Verbosity -transform=snake-upper
type ImageDownload string

const (
	Never   ImageDownload = "never"
	Always                = "always"
	Missing               = "missing"
)

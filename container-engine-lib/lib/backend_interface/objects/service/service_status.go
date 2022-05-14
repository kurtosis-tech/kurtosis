package service

//go:generate go run github.com/dmarkham/enumer -trimprefix=UserServiceStatus_ -transform=snake-upper -type=UserServiceStatus
type UserServiceStatus int
const (
	UserServiceStatus_Registered UserServiceStatus = iota
	UserServiceStatus_Running
	UserServiceStatus_Paused	// Indicates that the service has been paused, and all processes frozen
	UserServiceStatus_Stopped	// Indicates that the Service has been stopped and can no longer be used
)


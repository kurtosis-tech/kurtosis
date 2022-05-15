package service

/*
        Kurtosis Service State Diagram

REGISTERED ------------------------> DEACTIVATED
			 \                  /
			  '--> ACTIVATED --'
*/

//go:generate go run github.com/dmarkham/enumer -trimprefix=UserServiceStatus_ -transform=snake-upper -type=UserServiceStatus
type UserServiceStatus int
const (
	UserServiceStatus_Registered UserServiceStatus = iota	// A service does not have a container running
	UserServiceStatus_Activated			// Indicates that a service has a container started
	UserServiceStatus_Deactivated 		// Indicates that the Service can no longer be used (it may or may not have a container)
)


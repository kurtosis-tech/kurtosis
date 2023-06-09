package service

// Represents the state of a service within Kurtosis
//
//go:generate go run github.com/dmarkham/enumer -trimprefix=ServiceStatus_ -transform=snake-upper -type=ServiceStatus
type ServiceStatus int

const (
	ServiceStatus_Registered ServiceStatus = iota
	ServiceStatus_Started
	ServiceStatus_Stopped
)

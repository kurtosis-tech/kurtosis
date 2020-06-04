package services

import "time"

// TODO When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
type ServiceAvailabilityCheckerCore interface {
	// TODO When Go gets generics, make the type of these args 'toCheck S' and 'dependencies []N'
	IsServiceUp(toCheck Service, dependencies []Service) bool

	// How long to keep checking for the service to be available before giving up
	GetTimeout() time.Duration
}

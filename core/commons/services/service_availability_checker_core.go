package services

import "time"

// GENERICS TOOD: When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
/*
Provides the condition to dictate when Kurtosis should consider a service as available.
 */
type ServiceAvailabilityCheckerCore interface {
	// GENERICS TOOD: When Go gets generics, make the type of these args 'toCheck S' and 'dependencies []N'
	// Performs a service-specific check against the given service (and possibly its dependencies) to check if it's available
	IsServiceUp(toCheck Service, dependencies []Service) bool

	// How long to keep checking for the service to be available before giving up
	GetTimeout() time.Duration
}

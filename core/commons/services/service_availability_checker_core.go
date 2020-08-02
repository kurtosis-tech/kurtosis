package services

import "time"

// GENERICS TOOD: When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
/*
Provides the user-defined condition to dictate when Kurtosis should consider a service as available. The user of this
	library should create one core per *type* of service being used in a test network.
 */
type ServiceAvailabilityCheckerCore interface {
	// GENERICS TOOD: When Go gets generics, make the type of these args 'toCheck S' and 'dependencies []N'
	/*
	Performs a service-specific check against the given service (and possibly its dependencies) to check if it's available.

	Args:
		toCheck: The service to check. Because Go doesn't have generics, the user will need to cast this to the expected
			interface type.
		dependencies: The dependencies of the service to check, which are provided only in the event that they're needed

	Returns:
		True if the service is available, false if not
	 */
	IsServiceUp(toCheck Service, dependencies []Service) bool

	// How long to keep checking for the service to be available before giving up
	GetTimeout() time.Duration
}

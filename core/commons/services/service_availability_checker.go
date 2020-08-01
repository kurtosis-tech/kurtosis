package services

import (
	"context"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	TIME_BETWEEN_STARTUP_POLLS = 1 * time.Second
)

/*
Contains the logic wrapping a ServiceAvailabilityCheckerCore, which is used to make requests against a service and verify
	if it's actually available (because a Docker container running doesn't necessarily mean that the service is running).
	We have the notion of a user-defined "core" because the criteria for a service being available will depend on the
	developer's service.
 */
type ServiceAvailabilityChecker struct {
	/*
	The context that the availability checker is running in, which will be used for cancelling/timing out.

	NOTE: We bend the Go rules and store a context in a struct because we don't want the user to need to think about contexts
		when writing their tests
	 */
	context context.Context

	// The developer-defined criteria for determining if their custom service is available
	core ServiceAvailabilityCheckerCore

	// The service that will be checked for availability
	toCheck Service

	// The dependencies that the service-to-check depends on (just in case it's useful)
	dependencies []Service
}

/*
Creates a new availability checker using the given core.

Args:
	context: The context that availability-checking will happen in, which can be used to check availability checking
	core: The user-defined criteria for whether their custom service is up
	toCheck: The service to check
	dependencies: The dependencies of the service being checked
 */
func NewServiceAvailabilityChecker(context context.Context, core ServiceAvailabilityCheckerCore, toCheck Service, dependencies []Service) *ServiceAvailabilityChecker {
	// Defensive copy
	dependenciesCopy := make([]Service, 0, len(dependencies))
	copy(dependenciesCopy, dependencies)

	return &ServiceAvailabilityChecker{
		context: context,
		core: core,
		toCheck: toCheck,
		dependencies: dependenciesCopy,
	}
}

/*
Waits for the service that was passed in at construction time to start up by making requests to the service until
	the availability checker core's criteria are met or the timeout is reached.
 */
func (checker ServiceAvailabilityChecker) WaitForStartup() error {
	startupTimeout := checker.core.GetTimeout()

	timeoutContext, cancel := context.WithTimeout(checker.context, startupTimeout)
	defer cancel()

	for timeoutContext.Err() == nil {
		if checker.core.IsServiceUp(checker.toCheck, checker.dependencies) {
			return nil
		}
		logrus.Tracef("Service is not yet available; sleeping for %v before retrying...", TIME_BETWEEN_STARTUP_POLLS)
		time.Sleep(TIME_BETWEEN_STARTUP_POLLS)
	}

	contextErr := timeoutContext.Err()
	if (contextErr == context.Canceled) {
		return stacktrace.Propagate(contextErr, "Context was cancelled while waiting for service to startFailed to Hit timeout (%v) while waiting for service to start", startupTimeout)
	} else if (contextErr == context.DeadlineExceeded) {
		return stacktrace.Propagate(contextErr, "Hit timeout (%v) while waiting for service to start", startupTimeout)
	} else {
		return stacktrace.Propagate(contextErr, "Hit an unknown context error while waiting for service to start")
	}
}

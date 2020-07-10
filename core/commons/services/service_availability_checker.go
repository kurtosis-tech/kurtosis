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

type ServiceAvailabilityChecker struct {
	context context.Context // We bend the Go rules and store this context we don't want the user to need to think about it when writing their tests
	core ServiceAvailabilityCheckerCore
	toCheck Service
	dependencies []Service
}

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

// Waits for the linked service to start up by making requests (configured by the core) to the service until the service
//  is reported as up or the timeout is reached
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

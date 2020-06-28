package services

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	TIME_BETWEEN_STARTUP_POLLS = 1 * time.Second
)

type ServiceAvailabilityChecker struct {
	core ServiceAvailabilityCheckerCore
	toCheck Service
	dependencies []Service
}

func NewServiceAvailabilityChecker(core ServiceAvailabilityCheckerCore, toCheck Service, dependencies []Service) *ServiceAvailabilityChecker {
	// Defensive copy
	dependenciesCopy := make([]Service, 0, len(dependencies))
	copy(dependenciesCopy, dependencies)

	return &ServiceAvailabilityChecker{
		core: core,
		toCheck: toCheck,
		dependencies: dependenciesCopy,
	}
}

// Waits for the linked service to start up by making requests (configured by the core) to the service until the service
//  is reported as up or the timeout is reached
func (checker ServiceAvailabilityChecker) WaitForStartup() error {
	startupTimeout := checker.core.GetTimeout()
	pollStartTime := time.Now()
	for time.Since(pollStartTime) < startupTimeout {
		if checker.core.IsServiceUp(checker.toCheck, checker.dependencies) {
			return nil
		}
		logrus.Tracef("Service is not yet available; sleeping for %v before retrying...", TIME_BETWEEN_STARTUP_POLLS)
		time.Sleep(TIME_BETWEEN_STARTUP_POLLS)
	}
	return stacktrace.NewError("Hit timeout (%v) while waiting for service to start", startupTimeout)
}

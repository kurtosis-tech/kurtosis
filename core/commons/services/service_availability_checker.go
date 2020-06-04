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
}

func NewServiceAvailabilityChecker(core ServiceAvailabilityCheckerCore) *ServiceAvailabilityChecker {
	return &ServiceAvailabilityChecker{core: core}
}

// Waits for the given service to start up by making requests (configured by the core) to the service until the service
//  is reported as up or the timeout is reached
func (checker ServiceAvailabilityChecker) WaitForStartup(toCheck Service, dependencies []Service) error {
	startupTimeout := checker.core.GetTimeout()
	pollStartTime := time.Now()
	for time.Since(pollStartTime) < startupTimeout {
		if checker.core.IsServiceUp(toCheck, dependencies) {
			return nil
		}
		logrus.Tracef("Service is not yet available; sleeping for %v before retrying...", TIME_BETWEEN_STARTUP_POLLS)
		time.Sleep(TIME_BETWEEN_STARTUP_POLLS)
	}
	return stacktrace.NewError("Hit timeout (%v) while waiting for service to start", startupTimeout)
}

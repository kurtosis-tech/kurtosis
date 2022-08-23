package logs_components

type LogsCollectorAvailabilityChecker interface {
	WaitForAvailability() error
}

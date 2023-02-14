package logs_collector_functions

type LogsCollectorAvailabilityChecker interface {
	WaitForAvailability() error
}

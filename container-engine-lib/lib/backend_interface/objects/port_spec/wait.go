package port_spec

import (
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	DefaultWaitTimeoutDurationStr = "15s"
	DisableWaitTimeoutDurationStr = ""
)

type Wait struct {
	timeout time.Duration
}

func NewWait(timeout time.Duration) *Wait {
	return &Wait{timeout: timeout}
}

func CreateWaitWithDefaultValues() (*Wait, error) {
	defaultTimeout, err := time.ParseDuration(DefaultWaitTimeoutDurationStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new wait with default values")
	}

	newWait := &Wait{
		timeout: defaultTimeout,
	}

	return newWait, nil
}

func CreateWait(timeoutStr string) (*Wait, error) {

	timeoutDuration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a wait object using time out string '%s'", timeoutStr)
	}

	return &Wait{timeout: timeoutDuration}, nil
}

func (wait *Wait) GetTimeout() time.Duration {
	return wait.timeout
}

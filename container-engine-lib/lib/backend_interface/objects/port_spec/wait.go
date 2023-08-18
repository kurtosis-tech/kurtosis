package port_spec

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	DefaultWaitTimeoutDurationStr = "2m"
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

	newWait := NewWait(defaultTimeout)

	return newWait, nil
}

func CreateWait(timeoutStr string) (*Wait, error) {

	timeoutDuration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a wait object using time out string '%s'", timeoutStr)
	}

	newWait := NewWait(timeoutDuration)

	return newWait, nil
}

func (wait *Wait) GetTimeout() time.Duration {
	return wait.timeout
}

func (wait *Wait) String() string {
	return wait.timeout.String()
}

func (wait *Wait) MarshalJSON() ([]byte, error) {

	return json.Marshal(wait.timeout)
}

func (wait *Wait) UnmarshalJSON(data []byte) error {

	var timeout time.Duration

	timeOutPtr := &timeout

	if err := json.Unmarshal(data, timeOutPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling time duration")
	}

	wait.timeout = timeout
	return nil
}

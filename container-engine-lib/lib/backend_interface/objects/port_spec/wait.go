package port_spec

import "time"

const (
	defaultTimeout = 15 * time.Second
)

type wait struct {
	timeout time.Duration
}

func NewWait(timeout time.Duration) *wait {
	return &wait{timeout: timeout}
}

func newWaitWithDefaultValues() *wait {
	return &wait{
		timeout: defaultTimeout,
	}
}

func (wait *wait) GetTimeout() time.Duration {
	return wait.timeout
}

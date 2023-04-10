package port_spec

import "time"

const (
	enableByDefault         = true
	defaultTimeout          = 15 * time.Second
	noInitialDelayByDefault = time.Duration(0)
)

//TODO we probably will rename it, it's in the design stage
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

package port_spec

import "time"

const (
	enableByDefault         = true
	defaultTimeout          = 15 * time.Second
	noInitialDelayByDefault = time.Duration(0)
)

//TODO we probably will rename it, it's in the design stage
type wait struct {
	enable       bool
	timeout      time.Duration
	initialDelay time.Duration
}

func NewWait(enable bool, timeout time.Duration, initialDelay time.Duration) *wait {
	return &wait{enable: enable, timeout: timeout, initialDelay: initialDelay}
}

func newWaitWithDefaultValues() *wait {
	return &wait{
		enable:       enableByDefault,
		timeout:      defaultTimeout,
		initialDelay: noInitialDelayByDefault,
	}
}

func (wait *wait) GetEnable() bool {
	return wait.enable
}

func (wait *wait) GetTimeout() time.Duration {
	return wait.timeout
}

func (wait *wait) GetInitialDelay() time.Duration {
	return wait.initialDelay
}

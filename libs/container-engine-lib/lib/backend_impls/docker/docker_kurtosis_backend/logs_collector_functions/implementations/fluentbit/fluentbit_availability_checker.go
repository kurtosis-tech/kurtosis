package fluentbit

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"net/http"
	"time"
)

const (
	waitForAvailabilityInitialDelayMilliseconds = 100
	waitForAvailabilityMaxRetries               = 20
	waitForAvailabilityRetriesDelayMilliseconds = 50
)

type fluentbitAvailabilityChecker struct {
	ipAddr net.IP
	httpPortNumber uint16
}

func NewFluentbitAvailabilityChecker(ipAddr net.IP, httpPortNumber uint16) *fluentbitAvailabilityChecker {
	return &fluentbitAvailabilityChecker{ipAddr: ipAddr, httpPortNumber: httpPortNumber}
}

func (fluent *fluentbitAvailabilityChecker) WaitForAvailability() error {

	return waitForEndpointAvailability(
		fluent.ipAddr.String(),
		fluent.httpPortNumber,
		healthCheckEndpointPath,
		waitForAvailabilityInitialDelayMilliseconds,
		waitForAvailabilityMaxRetries,
		waitForAvailabilityRetriesDelayMilliseconds,
	)
}

func waitForEndpointAvailability(
	host string,
	port uint16,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
) error {

	var err error

	url := fmt.Sprintf("%v://%v:%v/%v", httpProtocolStr, host, port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		_, err = makeHttpRequest(url)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	return nil
}

func makeHttpRequest(url string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	resp, err = http.Get(url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v' ", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

package wait_for_availability_http_methods

// Represents the availables http methods that can be used in wait for http endpoint availability method
//go:generate go run github.com/dmarkham/enumer -trimprefix=WaitForAvailabilityHttpMethod_ -transform=snake-upper -type=WaitForAvailabilityHttpMethod
type WaitForAvailabilityHttpMethod int
const (
	WaitForAvailabilityHttpMethod_GET WaitForAvailabilityHttpMethod = iota

	WaitForAvailabilityHttpMethod_POST

	WaitForAvailabilityHttpMethod_HEAD
)

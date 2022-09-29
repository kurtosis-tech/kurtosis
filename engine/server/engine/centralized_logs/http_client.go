package centralized_logs

import "net/http"

// httpClient is an interface for testing a request object.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

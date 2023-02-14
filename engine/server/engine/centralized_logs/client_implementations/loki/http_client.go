package loki

import "net/http"

// httpClient is an interface for testing purpose.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

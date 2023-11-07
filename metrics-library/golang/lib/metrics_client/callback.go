package metrics_client

// Callback methods must return quickly and not cause long blocking operations
// to avoid interfering with the client's internal work flow.
type Callback interface {

	// This method is called for every message that was successfully sent
	Success()

	// This method is called for every message that failed to be sent
	// and will be discarded by the client.
	Failure(error)
}

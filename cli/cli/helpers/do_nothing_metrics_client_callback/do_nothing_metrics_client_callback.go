package do_nothing_metrics_client_callback

// this call back does nothing
// the metrics client allows us to specify a call back that does things after a successful/failure
// we use it in the install-consent to clear the backlog & associated file; but we don't use it for other metrics
type doNothingMetricsClientCallback struct{}

func NewDoNothingMetricsClientCallback() doNothingMetricsClientCallback {
	return doNothingMetricsClientCallback{}
}

func (d doNothingMetricsClientCallback) Success()          {}
func (d doNothingMetricsClientCallback) Failure(err error) {}

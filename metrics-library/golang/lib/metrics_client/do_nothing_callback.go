package metrics_client

type DoNothingMetricsClientCallback struct{}

func (d DoNothingMetricsClientCallback) Success()          {}
func (d DoNothingMetricsClientCallback) Failure(err error) {}

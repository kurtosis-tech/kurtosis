package client

const (
	Segment MetricsClientType = "segment"
	//It's used when users reject sending metrics
	DoNothing MetricsClientType = "do-nothing"
)

type MetricsClientType string

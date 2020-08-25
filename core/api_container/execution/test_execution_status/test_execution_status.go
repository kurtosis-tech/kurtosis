package test_execution_status

type TestExecutionStatus string

const (
	Running                TestExecutionStatus = "RUNNING"
	CompletedBeforeTimeout TestExecutionStatus = "COMPLETED_BEFORE_TIMEOUT"
	HitTimeout             TestExecutionStatus = "HIT_TIMEOUT"
)

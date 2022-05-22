module github.com/kurtosis-tech/kurtosis-core/files_artifact_expander

replace (
	github.com/kurtosis-tech/kurtosis-core/api/golang => ../api/golang
)

require (
	github.com/kurtosis-tech/kurtosis-core/api/golang v0.0.0
	github.com/gammazero/workerpool v1.1.2
)
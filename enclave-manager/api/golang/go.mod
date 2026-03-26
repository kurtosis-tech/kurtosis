module github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang

go 1.26.0

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../../api/golang
	github.com/kurtosis-tech/kurtosis/cloud/api/golang => ../../../cloud/api/golang
)

require (
	connectrpc.com/connect v1.19.1
	github.com/kurtosis-tech/kurtosis/api/golang v0.81.9
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.88.12
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.11
)

require (
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.40.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

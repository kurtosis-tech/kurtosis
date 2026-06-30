module github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang

go 1.26.0

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../../api/golang
	github.com/kurtosis-tech/kurtosis/cloud/api/golang => ../../../cloud/api/golang
)

require (
	connectrpc.com/connect v1.20.0
	github.com/kurtosis-tech/kurtosis/api/golang v0.81.9
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.88.12
	google.golang.org/grpc v1.82.0
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
)

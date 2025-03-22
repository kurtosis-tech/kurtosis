module github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang

go 1.20

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../../api/golang
	github.com/kurtosis-tech/kurtosis/cloud/api/golang => ../../../cloud/api/golang
)

require (
	connectrpc.com/connect v1.11.1
	github.com/kurtosis-tech/kurtosis/api/golang v0.81.9
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.88.12
	google.golang.org/grpc v1.57.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230803162519-f966b187b2e5 // indirect
)

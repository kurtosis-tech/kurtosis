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
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
)

require (
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
)

replace github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../../contexts-config-store

replace github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../../grpc-file-transfer/golang

replace github.com/kurtosis-tech/kurtosis/path-compression => ../../../path-compression

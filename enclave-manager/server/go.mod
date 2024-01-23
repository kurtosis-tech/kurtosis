module github.com/kurtosis-tech/kurtosis/enclave-manager

go 1.20

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/cloud/api/golang => ../../cloud/api/golang
	github.com/kurtosis-tech/kurtosis/connect-server => ../../connect-server
	github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang => ../api/golang
)

require (
	connectrpc.com/connect v1.11.1
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/connect-server v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/rs/cors v1.9.0
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230706204954-ccb25ca9f130 // indirect
	google.golang.org/grpc v1.56.3 // indirect
)

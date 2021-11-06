module github.com/kurtosis-tech/kurtosis-engine-server

go 1.15

require (
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211106215242-31501d8d7f8d
	github.com/kurtosis-tech/kurtosis-client/golang v0.0.0-20211027222420-ebca40d7f918
	github.com/kurtosis-tech/kurtosis-core v0.0.0-20211103233136-3dbd50327378
	github.com/kurtosis-tech/kurtosis-engine-api-lib/golang v0.0.0-20211101165721-7075d4829152
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

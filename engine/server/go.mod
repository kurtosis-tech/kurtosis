module github.com/kurtosis-tech/kurtosis-engine-server/server

go 1.15

replace (
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang => ../api/golang
	github.com/kurtosis-tech/kurtosis-engine-server/launcher => ../launcher
)

require (
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211106215243-ccb878a45a90
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-d3be9e82993e
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20211109204911-1371c050a5d8 // indirect
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20211109204724-e5d181ea0fb3
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20211109200015-fa88b22da2d8
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

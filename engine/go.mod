module github.com/kurtosis-tech/kurtosis-engine-server

go 1.15

require (
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/docker/docker v20.10.9+incompatible
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211013224800-47b5d7199d68
	github.com/kurtosis-tech/kurtosis-core v0.0.0-20211013195011-74f1fcb1bee6
	github.com/kurtosis-tech/kurtosis-engine-api-lib/golang v0.0.0-20211014185242-b5cc89c705a7
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
)

module github.com/kurtosis-tech/kurtosis-engine-server

go 1.15

require (
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/hashicorp/go-retryablehttp v0.6.7 // indirect
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211013191503-b01ed3982dcd
	github.com/kurtosis-tech/example-microservice v0.0.0-20210708190343-51d08a1c685b // indirect
	github.com/kurtosis-tech/kurtosis-core v0.0.0-20211018165703-86ed1577a258
	github.com/kurtosis-tech/kurtosis-engine-api-lib/golang v0.0.0-20211019002658-fc6ee53f0cb0 // TDOOOOOO Make this a proper release version!!!
	github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang v0.0.0-20210719191644-4ea5a878960c // indirect
	github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang v0.0.0-20210924144702-dac0cddbd2b1 // indirect
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

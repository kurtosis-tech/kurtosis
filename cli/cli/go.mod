module github.com/kurtosis-tech/kurtosis-cli/cli

go 1.15

replace github.com/kurtosis-tech/kurtosis-cli/commons => ../commons

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/adrg/xdg v0.4.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211116225347-a5bd1c49b423
	github.com/kurtosis-tech/kurtosis-cli/commons v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20211116225622-72737e294f73
	github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang v0.0.0-20211117013451-7960199bfb27
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0-20211117013435-51e7caa6a6e4
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20211111005229-ffdab131bcda
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

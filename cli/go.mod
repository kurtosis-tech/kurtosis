module github.com/kurtosis-tech/kurtosis-cli

go 1.15

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211007193409-610c9d2e2a45
	github.com/kurtosis-tech/example-microservice v0.0.0-20210708190343-51d08a1c685b
	github.com/kurtosis-tech/kurtosis-client/golang v0.0.0-20211005195321-3134b69ea2c6
	github.com/kurtosis-tech/kurtosis-core v0.0.0-20211012200027-e18666699991
	github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang v0.0.0-20210924144702-dac0cddbd2b1
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
)

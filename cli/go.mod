module github.com/kurtosis-tech/kurtosis-cli

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/adrg/xdg v0.4.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/golang/protobuf v1.5.2
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211106215242-31501d8d7f8d
	github.com/kurtosis-tech/example-api-server/api/golang v0.0.0-20211101152411-a56fef9e73dd
	github.com/kurtosis-tech/example-datastore-server/api/golang v0.0.0-20211101145825-570cf60ea641
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20211104015055-c5e7cefde0aa
	github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang v0.0.0-20211107211522-ef09c391d535
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
)

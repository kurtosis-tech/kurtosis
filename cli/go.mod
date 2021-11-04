module github.com/kurtosis-tech/kurtosis-cli

go 1.15

require (
	github.com/adrg/xdg v0.4.0
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211103232750-85edc2a39f9e
	github.com/kurtosis-tech/example-api-server/api/golang v0.0.0-20211101152411-a56fef9e73dd
	github.com/kurtosis-tech/example-datastore-server/api/golang v0.0.0-20211101145825-570cf60ea641
	github.com/kurtosis-tech/kurtosis-client/golang v0.0.0-20211027222420-ebca40d7f918
	github.com/kurtosis-tech/kurtosis-core v0.0.0-20211103233136-78a97d6bef99
	github.com/kurtosis-tech/kurtosis-engine-api-lib/golang v0.0.0-20211101165721-7075d4829152
	github.com/kurtosis-tech/kurtosis-engine-server v0.0.0-20211103234101-61519bf53ebf
	github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang v0.0.0-20211027222833-7233d903873e
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b // indirect
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
)

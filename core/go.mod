// TODO Refactor last package to be kurtosis-core!
module github.com/kurtosis-tech/kurtosis

go 1.13

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20210828221519-094b8e2cd627
	github.com/kurtosis-tech/example-microservice v0.0.0-20210708190343-51d08a1c685b
	github.com/kurtosis-tech/kurtosis-client/golang v0.0.0-20210906125957-f783be61d52b
	github.com/kurtosis-tech/kurtosis-core-launcher-lib v0.0.0-20210904011652-624b6d90bdaf
	github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang v0.0.0-20210719191644-4ea5a878960c
	github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang v0.0.0-20210803203144-801fd9867a7f
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210821054858-4cac5a245c85
	github.com/moby/term v0.0.0-20200507201656-73f35e472e8f // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
)

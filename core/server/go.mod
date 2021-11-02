module github.com/kurtosis-tech/kurtosis-core/server

go 1.13

replace github.com/kurtosis-tech/kurtosis-core/api/golang => ../api/golang

require (
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/martian v2.1.0+incompatible
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211021172205-bf12424cc95d
	github.com/kurtosis-tech/kurtosis-core/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-module-api-lib/golang v0.0.0-20211027222830-d8f7dfe68c3e
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20210921153930-d70d7667c51b
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/moby/term v0.0.0-20200507201656-73f35e472e8f // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gotest.tools v2.2.0+incompatible
)

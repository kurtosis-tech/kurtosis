module github.com/kurtosis-tech/kurtosis-core/server

go 1.13

replace (
	github.com/kurtosis-tech/kurtosis-core/api/golang => ../api/golang
	github.com/kurtosis-tech/kurtosis-core/launcher => ../launcher
)

require (
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211106215242-31501d8d7f8d
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-d3be9e82993e
	github.com/kurtosis-tech/kurtosis-client/golang v0.0.0-20211027222420-ebca40d7f918
	github.com/kurtosis-tech/kurtosis-core/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0
	github.com/kurtosis-tech/kurtosis-module-api-lib/golang v0.0.0-20211027222830-d8f7dfe68c3e
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211108000544-89f5ac8d0f88
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20211104234832-e10434db185a
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/moby/term v0.0.0-20200507201656-73f35e472e8f // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gotest.tools v2.2.0+incompatible
)

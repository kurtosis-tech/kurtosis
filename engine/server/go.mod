module github.com/kurtosis-tech/kurtosis/engine/server

go 1.18

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../contexts-config-store
	github.com/kurtosis-tech/kurtosis/core/launcher => ../../core/launcher
	github.com/kurtosis-tech/kurtosis/engine/launcher => ../launcher
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
	github.com/kurtosis-tech/kurtosis/name_generator => ../../name_generator
)

require (
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/core/launcher v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/engine/launcher v0.0.0
	github.com/kurtosis-tech/kurtosis/name_generator v0.0.0 // local dependency
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20230120102313-4adf4f054260
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.29.1
)

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v20.10.24+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/gorilla/websocket v1.4.2

require (
	github.com/kurtosis-tech/kurtosis/kurtosis_version v0.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)

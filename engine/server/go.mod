module github.com/kurtosis-tech/kurtosis-engine-server/server

go 1.15

replace (
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang => ../api/golang
	github.com/kurtosis-tech/kurtosis-engine-server/launcher => ../launcher
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220426170434-b6da6d148295
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-d3be9e82993e // indirect
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20220428115252-6668f26c7fdd // indirect
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20220428115230-e66d7a4e56af
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20220215151652-4f1a58645739
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211201000847-a204edc5a0b3
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220225193403-74da3f3b98ce
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

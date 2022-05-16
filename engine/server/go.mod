module github.com/kurtosis-tech/kurtosis-engine-server/server

go 1.15

replace (
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang => ../api/golang
	github.com/kurtosis-tech/kurtosis-engine-server/launcher => ../launcher
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220516025347-10fa2badb2d7
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20220516025634-08b3e63ca8ef // indirect
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20220516025619-39ed776e35f9
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20220215151652-4f1a58645739
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211201000847-a204edc5a0b3
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220225193403-74da3f3b98ce
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

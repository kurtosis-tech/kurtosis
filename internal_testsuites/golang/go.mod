module github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite

go 1.19

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
)

require (
	github.com/golang/protobuf v1.5.3
	github.com/kurtosis-tech/example-api-server/api/golang v0.0.0-20211207020812-00a54fc29318
	github.com/kurtosis-tech/example-datastore-server/api/golang v0.0.0-20211207020830-504dbf5ed1a6
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // local dependency
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.3
	google.golang.org/grpc v1.56.1
	google.golang.org/protobuf v1.31.0
)

require k8s.io/utils v0.0.0-20230711102312-30195339c3c7

require (
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kurtosis-tech/kurtosis-portal/api/golang v0.0.0-20230328194643-b4dea3081e25 // indirect
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang v0.0.0 // indirect
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230629202037-9506855d4529 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

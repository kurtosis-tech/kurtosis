module github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite

go 1.18

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../contexts-config-store
)

require (
	github.com/golang/protobuf v1.5.2
	github.com/kurtosis-tech/example-api-server/api/golang v0.0.0-20211207020812-00a54fc29318
	github.com/kurtosis-tech/example-datastore-server/api/golang v0.0.0-20211207020830-504dbf5ed1a6
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // local dependency
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.4
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kurtosis-tech/kurtosis-portal/api/golang v0.0.0-20230328194643-b4dea3081e25 // indirect
	github.com/kurtosis-tech/kurtosis/contexts-config-store v0.0.0 // indirect
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

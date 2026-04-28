module github.com/kurtosis-tech/kurtosis/api/golang

go 1.26.0

replace (
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
	github.com/kurtosis-tech/kurtosis/path-compression => ../../path-compression
)

require (
	connectrpc.com/connect v1.19.1
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/getkin/kin-openapi v0.135.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kurtosis-tech/kurtosis-portal/api/golang v0.0.0-20230818182330-1a86869414d2
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.0.0-20230803130419-099ee7a4e3dc
	github.com/kurtosis-tech/kurtosis/contexts-config-store v0.0.0-20230818184218-f4e3e773463b
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang v0.0.0-20230803130419-099ee7a4e3dc // needs to be pinned as the replace above won't work when importing the api standalone
	github.com/kurtosis-tech/kurtosis/path-compression v0.0.0-20260325155815-f36ae687d73d // needs to be pinned as the replace above won't work when importing the api standalone
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/labstack/echo/v4 v4.15.1
	github.com/oapi-codegen/runtime v1.4.0
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mholt/archives v0.1.5 // indirect
	github.com/mikelolasagasti/xz v1.0.1 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/nwaples/rardecode/v2 v2.2.0 // indirect
	github.com/oasdiff/yaml v0.0.9 // indirect
	github.com/oasdiff/yaml3 v0.0.9 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/woodsbury/decimal128 v1.3.0 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

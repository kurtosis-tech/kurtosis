module github.com/kurtosis-tech/kurtosis/api/golang

go 1.23.3

toolchain go1.23.7

replace (
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
	github.com/kurtosis-tech/kurtosis/path-compression => ../../path-compression
)

require (
	connectrpc.com/connect v1.11.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/getkin/kin-openapi v0.120.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kurtosis-tech/kurtosis-portal/api/golang v0.0.0-20230818182330-1a86869414d2
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.0.0-20230803130419-099ee7a4e3dc
	github.com/kurtosis-tech/kurtosis/contexts-config-store v0.0.0-20230818184218-f4e3e773463b
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang v0.0.0-20230803130419-099ee7a4e3dc // needs to be pinned as the replace above won't work when importing the api standalone
	github.com/kurtosis-tech/kurtosis/path-compression v0.0.0-20240307154559-64d2929cd265 // needs to be pinned as the replace above won't work when importing the api standalone
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/labstack/echo/v4 v4.11.3
	github.com/oapi-codegen/runtime v1.1.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.57.1
	google.golang.org/protobuf v1.34.1
)

require (
	github.com/STARRY-S/zip v0.2.2 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.0 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/jm33-m0/arc/v2 v2.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mholt/archives v0.1.1 // indirect
	github.com/minio/minlz v1.0.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/nwaples/rardecode/v2 v2.1.0 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sorairolake/lzip-go v0.3.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto v0.0.0-20230726155614-23370e0ffb3e // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230803162519-f966b187b2e5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

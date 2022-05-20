module github.com/kurtosis-tech/kurtosis-cli/cli

go 1.17

replace github.com/kurtosis-tech/kurtosis-cli/commons => ../commons

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/adrg/xdg v0.4.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dmarkham/enumer v1.5.5
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.16+incompatible
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220520170436-c1b29432bb40
	github.com/kurtosis-tech/kurtosis-cli/commons v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20220520173359-57be80203ed3
	github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang v0.0.0-20220520174311-571d508e9d48
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0-20220520174253-c57959c232d4
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20220215151652-4f1a58645739
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211205213337-f5088fc26465
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220225193403-74da3f3b98ce
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.14
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
)

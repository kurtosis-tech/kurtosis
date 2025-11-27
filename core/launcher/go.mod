module github.com/dzobbe/PoTE-kurtosis/core/launcher

go 1.23.0

toolchain go1.23.7

replace (
	github.com/dzobbe/PoTE-kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/dzobbe/PoTE-kurtosis/kurtosis_version => ../../kurtosis_version
)

require (
	github.com/dzobbe/PoTE-kurtosis/container-engine-lib v0.0.0 // Local dependency
	github.com/dzobbe/PoTE-kurtosis/kurtosis_version v0.0.0 // Local dependency generated during build
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
)

require github.com/dzobbe/PoTE-kurtosis/metrics-library/golang v0.0.0-20231206095907-9bdf0d02cb90

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/segmentio/analytics-go.v3 v3.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.27.2 // indirect
	k8s.io/apimachinery v0.27.2 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

module github.com/kurtosis-tech/kurtosis/engine/launcher

go 1.20

replace (
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
	github.com/kurtosis-tech/kurtosis/metrics-library/golang => ../../metrics-library/golang
)

require (
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/kurtosis_version v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/metrics-library/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/sys v0.15.0 // indirect
	gopkg.in/segmentio/analytics-go.v3 v3.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.27.2 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
)

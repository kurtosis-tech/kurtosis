module github.com/kurtosis-tech/container-engine-lib

go 1.15

require (
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0-20220208161434-a0488f5d78fd
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220207150232-548c80e05196
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

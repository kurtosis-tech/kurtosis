module github.com/kurtosis-tech/container-engine-lib

go 1.15

require (
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220207150232-90a0b5257ee2
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e // indirect
	google.golang.org/grpc v1.40.0 // indirect
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

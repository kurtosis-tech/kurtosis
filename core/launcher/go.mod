module github.com/kurtosis-tech/kurtosis/core/launcher

go 1.26.0

replace (
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
)

require (
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.10.0
	github.com/stretchr/testify v1.12.0
)

require (
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0-00010101000000-000000000000
	github.com/kurtosis-tech/kurtosis/kurtosis_version v0.0.0-00010101000000-000000000000
	github.com/kurtosis-tech/kurtosis/metrics-library/golang v0.0.0-20231206095907-9bdf0d02cb90
)

require (
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/segmentio/backo-go v1.1.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/segmentio/analytics-go.v3 v3.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.36.1 // indirect
	k8s.io/apimachinery v0.36.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
)

replace github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang

replace github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../contexts-config-store

replace github.com/kurtosis-tech/kurtosis/engine/launcher => ../../engine/launcher

replace github.com/kurtosis-tech/kurtosis/metrics-library/golang => ../../metrics-library/golang

replace github.com/kurtosis-tech/kurtosis/path-compression => ../../path-compression

module github.com/kurtosis-tech/kurtosis/engine/server

go 1.19

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/connect-server => ../../connect-server
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../contexts-config-store
	github.com/kurtosis-tech/kurtosis/core/launcher => ../../core/launcher
	github.com/kurtosis-tech/kurtosis/enclave-manager/server => ../../enclave-manager/server
	github.com/kurtosis-tech/kurtosis/engine/launcher => ../launcher
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
	github.com/kurtosis-tech/kurtosis/name_generator => ../../name_generator
)

require (
	github.com/kurtosis-tech/kurtosis/api/golang v0.84.10
	github.com/kurtosis-tech/kurtosis/connect-server v0.0.0
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/core/launcher v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/engine/launcher v0.0.0
	github.com/kurtosis-tech/kurtosis/name_generator v0.0.0 // local dependency
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.56.2 // indirect
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.7+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	connectrpc.com/connect v1.11.1
	github.com/kurtosis-tech/kurtosis/enclave-manager/server v0.0.0-00010101000000-000000000000
	github.com/nxadm/tail v1.4.8
	github.com/rs/cors v1.9.0
	github.com/spf13/afero v1.10.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
	k8s.io/apimachinery v0.27.2
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kurtosis-tech/kurtosis/cloud/api/golang v0.0.0-20230828153722-32770ca96513 // indirect
	github.com/kurtosis-tech/kurtosis/contexts-config-store v0.0.0 // indirect
	github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang v0.0.0-20230828153722-32770ca96513 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230706204954-ccb25ca9f130 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.27.2 // indirect
	k8s.io/client-go v0.27.2 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

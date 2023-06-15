module github.com/kurtosis-tech/kurtosis/core/server

go 1.19

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander => ../files_artifacts_expander
	github.com/kurtosis-tech/kurtosis/core/launcher => ../launcher
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang => ../../grpc-file-transfer/golang
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
	github.com/kurtosis-tech/kurtosis/name_generator => ../../name_generator
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/core/launcher v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/name_generator v0.0.0 // Local dependency
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20230317105020-7ca453c242bd
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.29.1
	gotest.tools v2.2.0+incompatible
)

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/itchyny/gojq v0.12.9
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c
	go.etcd.io/bbolt v1.3.6
	go.starlark.net v0.0.0-20230224151120-c52844e64a10
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
	golang.org/x/sync v0.1.0
)

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v20.10.24+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/itchyny/timefmt-go v0.1.4 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.27.2 // indirect
	k8s.io/apimachinery v0.27.2 // indirect
	k8s.io/client-go v0.27.2 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230209194617-a36077c30491 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

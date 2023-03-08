module github.com/kurtosis-tech/kurtosis/cli/cli

go 1.18

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/engine/launcher => ../../engine/launcher
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/adrg/xdg v0.4.0
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dmarkham/enumer v1.5.5
	github.com/docker/distribution v2.7.1+incompatible
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/engine/launcher v0.0.0 // local dependency
	github.com/kurtosis-tech/kurtosis/kurtosis_version v0.0.0 // Local dependency generated during build
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20230221115618-70c305416224
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.14
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.6.1-0.20230225213037-567ea8ebc9b4
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.4
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	k8s.io/apimachinery v0.24.0 // indirect
	k8s.io/client-go v0.24.0
)

require github.com/bazelbuild/buildtools v0.0.0-20221110131218-762712d8ce3f

require (
	github.com/briandowns/spinner v1.20.0
	github.com/fatih/color v1.7.0
	github.com/mholt/archiver v3.1.1+incompatible
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v20.10.16+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-1f73d028840d // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pascaldekloe/name v1.0.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/term v0.1.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/segmentio/analytics-go.v3 v3.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

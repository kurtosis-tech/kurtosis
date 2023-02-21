module github.com/kurtosis-tech/kurtosis/core/server

go 1.18

replace (
	github.com/kurtosis-tech/kurtosis/api/golang => ../../api/golang
	github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
	github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander => ../files_artifacts_expander
	github.com/kurtosis-tech/kurtosis/core/launcher => ../launcher
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/kurtosis-tech/kurtosis/api/golang v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/container-engine-lib v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis/core/launcher v0.0.0 // Local dependency
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211201000847-a204edc5a0b3
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.4
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gotest.tools v2.2.0+incompatible
)

require (
	github.com/cenkalti/backoff/v4 v4.2.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/itchyny/gojq v0.12.9
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20230221115618-70c305416224
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.3.6
	go.starlark.net v0.0.0-20220928063852-5fccb4daaf6d
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.16+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/itchyny/timefmt-go v0.1.4 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-d3be9e82993e // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/segmentio/analytics-go.v3 v3.1.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

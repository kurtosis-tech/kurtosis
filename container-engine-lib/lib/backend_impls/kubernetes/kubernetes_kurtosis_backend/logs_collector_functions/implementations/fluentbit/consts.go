package fluentbit

const (
	healthCheckEndpointPath = "api/v1/health"

	fluentBitContainerName = "fluent-bit"
	// using debug image for now with testing toolkit (curl, netstat) eventually will move to latest regular img
	fluentBitImage = "fluent/fluent-bit:latest-debug"

	// volumes pulled from official fluent bit helm chart: https://github.com/fluent/helm-charts/blob/main/charts/fluent-bit/values.yaml
	varLogVolumeName                 = "varlog"
	varLogMountPath                  = "/var/log"
	varLibDockerContainersVolumeName = "varlibdockercontainers"
	varLibDockerContainersMountPath  = "/var/lib/docker/containers"
	varLogDockerContainersVolumeName = "varlogcontainers"
	varLogDockerContainersMountPath  = "/var/log/containers"

	fluentBitConfigVolumeName = "fluent-bit-config"
	fluentBitConfigMountPath  = "/fluent-bit/etc/conf"

	// for now, fluent bit will also stores all combined logs in files on the node
	// TODO: remove when output is logs aggregator
	fluentBitHostLogsVolumeName = "fluent-bit-host-logs"
	fluentBitHostLogsMountPath  = "/var/log/fluent-bit"

	// this db will store information about the offsets of log files the fluent bit log collector has processed
	// this way if it is restarted, fluent bit will start reading from where it left off and no logs will be missed
	// https://docs.fluentbit.io/manual/pipeline/inputs/tail#keep_state
	fluentBitCheckpointDbVolumeName = "fluent-bit-db"
	fluentBitCheckpointDbMountPath  = "/var/log/fluent-bit/db"

	// assuming this as default k8s api server url - this might not be the case for very custom k8s environments so making this a variable
	// in case it needs to be configured by the user down the line
	k8sApiServerUrl = "https://kubernetes.default.svc:443"

	// TODO: construct fluentbit config via go templating based on inputs
	fluentBitConfigFileName = "fluent-bit.conf"
	fluentBitConfigFmtStr   = `
[SERVICE]
    HTTP_Server       On
    HTTP_Listen       0.0.0.0
    HTTP_PORT         9713
    Parsers_File      /fluent-bit/etc/parsers.conf

[INPUT]
    Name              tail
    Tag               kurtosis.*
    Path              /var/log/containers/*_kt-*_user-service-container-*.log
    Parser            docker
    DB                /var/log/fluent-bit/db/fluent-bit.db
    DB.sync           normal
    Read_from_Head    true

[FILTER]
    Name              kubernetes
    Match             *
    Kube_URL          %v
    Merge_log         On
    Keep_Log          On
    Annotations       Off
    Labels            On

[OUTPUT]
    Name              stdout
    Match             *
    Format            json_lines

[OUTPUT]
    Name              file
    Match             *
    Path              /var/log/fluent-bit
    File              fluent-bit-output.log
    Format            plain
`
)

package fluentbit

const (
	healthCheckEndpointPath = "api/v1/health"

	fluentBitContainerName = "fluent-bit"
	fluentBitImage         = "fluent/fluent-bit:latest-debug"

	// volumes pulled from official fluent bit helm chart: https://github.com/fluent/helm-charts/blob/main/charts/fluent-bit/values.yaml
	varLogVolumeName                 = "varlog"
	varLogMountPath                  = "/var/log"
	varLibDockerContainersVolumeName = "varlibdockercontainers"
	varLibDockerContainersMountPath  = "/var/lib/docker/containers"
	varLogDockerContainersVolumeName = "varlogcontainers"
	varLogDockerContainersMountPath  = "/var/log/containers"

	fluentBit
	fluentBitConfigVolumeName = "fluent-bit-config"
	fluentBitConfigMountPath  = "/fluent-bit/etc/conf"

	// for now, fluent bit will also stores all combined logs in files on the node
	// TODO: remove when output is logs aggregator
	fluentBitHostLogsVolumeName = "fluent-bit-host-logs"
	fluentBitHostLogsMountPath  = "/avr/lfluent-bit-host-logsog/fluentbit"

	// TODO: construct fluentbit config via go templating based on inputs
	fluentBitConfigFileName = "fluent-bit.conf"
	fluentBitConfigStr      = `
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

[OUTPUT]
    Name              stdout
    Match             *
    Format            json_lines

[OUTPUT]
    Name              file
    Match             *
    Path              /fluent-bit-logs/
    File              fluentbit-output.log
    Format            plain

[FILTER]
    Name              kubernetes
    Match             kurtosis.*
    Merge_Log         On
    Merge_Log_Key     On
    Labels            On
    Annotations       On
    Kube_Tag_Prefix   kurtosis.var.log.containers.
`
)

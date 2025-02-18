package fluentbit

const (
	healthCheckEndpointPath = "api/v1/health"

	fluentBitImage = "fluent/fluent-bit:latest-debug"

	//logLevel               = "debug"
	//httpServerEnabledValue = "On"
	//httpServerLocalhost    = "0.0.0.0"

	fluentBitConfigStr = `
[SERVICE]
    HTTP_Server       On
    HTTP_Listen       0.0.0.0
    HTTP_PORT         2020
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

package vector

const (
	vectorContainerName = "vector"
	vectorImage         = "timberio/vector:0.31.0-debian"

	vectorConfigVolumeName = "vector-config"
	vectorConfigMountPath  = "/etc/vector"

	kurtosisLogsVolumeName = "varlogskurtosis"
	kurtosisLogsMountPath  = "/var/log/kurtosis"

	vectorConfigFileName = "vector.toml"
	vectorConfigFmtStr   = `
    data_dir = "/vector-data-dir"

    [sources.fluentbit]
    type = "fluent"
    address = "0.0.0.0:%v"

    [sinks.file_sink]
    type = "file"
    inputs = ["fluentbit"]
    path = "%v/%%G/%%V/{{ enclave_uuid }}/{{ service_uuid }}.json"

    [sinks.file_sink.encoding]
    codec = "json"
    
    [sinks.stdout_sink]
    type = "console"
    inputs = ["fluentbit"]
    target = "stdout"

    [sinks.stdout_sink.encoding]
    codec = "json"
`
)

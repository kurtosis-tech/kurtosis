package vector

const (
	vectorContainerName = "vector"
	vectorImage         = "timberio/vector:0.45.0-debian"

	vectorConfigVolumeName = "vector-config"
	vectorConfigMountPath  = "/etc/vector"

	kurtosisLogsVolumeName = "varlogskurtosis"
	kurtosisLogsMountPath  = "/var/log/kurtosis"

	apiPort = 8686

	// mount the data directory as the disk buffer for file sink is contained here and needs to be persisted onto the k8s node in case vector restarts
	vectorDataDirVolumeName = "varlibvector"
	vectorDataDirMountPath  = "/var/lib/vector"

	bufferSizeStr = "268435488" // 256 MB is min for vector

	vectorConfigFileName = "vector.toml"
	vectorConfigFmtStr   = `
    data_dir = "%v"

    [api]
    enabled = true
    address = "0.0.0.0:%v"

    [sources.fluentbit]
    type = "fluent"
    address = "0.0.0.0:%v"

    [sinks.file_sink]
    type = "file"
    inputs = ["fluentbit"]
    path = "%v/%%G/%%V/{{ %v }}/{{ %v }}.json"
   
    [sinks.file_sink.buffer]
	type = "disk"
	max_size = %v
	when_full = "block"

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

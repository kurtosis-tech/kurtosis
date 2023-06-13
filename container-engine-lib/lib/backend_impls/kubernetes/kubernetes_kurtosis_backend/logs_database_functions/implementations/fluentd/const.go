package fluentd

const (
	containerImage = "fluent/fluentd:stable"

	dirpath = "/logs/"

	configFileDir  = "/fluentd/etc/"
	configFileName = "fluent.conf"

	configFileTemplateName = "fluentdConfigFileTemplate"
	configFileTemplate     = `<source>
  @type forward
  port {{.PortNumber}}
  bind 0.0.0.0
</source>

<match **>
  @type file
  path ` + dirpath + `data.*.log
  append true
</match>
`

	configMapName      = "fluentd-config"
	fluentdName        = "fluentd"
	kubernetesAppLabel = "k8s-app"
)

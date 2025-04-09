package grafloki

const (
	lokiImage = "grafana/loki:3.4.2"
	lokiPort  = 3100

	grafanaImage = "grafana/grafana:11.6.0"
	grafanaPort  = 3000

	grafanaAuthAnonymousEnabledEnvVarKey   = "GF_AUTH_ANONYMOUS_ENABLED"
	grafanaAuthAnonymousEnabledEnvVarVal   = "true"
	grafanaAuthAnonymousOrgRoleEnvVarKey   = "GF_AUTH_ANONYMOUS_ORG_ROLE"
	grafanaAuthAnonymousOrgRoleEnvVarVal   = "Admin"
	grafanaSecurityAllowEmbeddingEnvVarKey = "GF_SECURITY_ALLOW_EMBEDDING"
	grafanaSecurityAllowEmbeddingEnvVarVal = "true"

	grafanaDatasourcesKey  = "datasources"
	grafanaDatasourcesPath = "/etc/grafana/provisioning/datasources"
)

type GrafanaDatasource struct {
	Name      string `yaml:"name"`
	Type_     string `yaml:"type"`
	Access    string `yaml:"access"`
	Url       string `yaml:"url"`
	IsDefault bool   `yaml:"isDefault"`
	Editable  bool   `yaml:"editable"`
}

type GrafanaDatasources struct {
	ApiVersion  int64               `yaml:"apiVersion"`
	Datasources []GrafanaDatasource `yaml:"datasources"`
}

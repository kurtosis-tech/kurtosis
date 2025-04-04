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

	grafanaDatasourcesKey = "datasources"
)

type GrafanaDatasource struct {
	name      string `yaml:"string"`
	type_     string `yaml:"type"`
	access    string `yaml:"access"`
	url       string `yaml:"url"`
	isDefault bool   `yaml:"isDefault"`
	editable  bool   `yaml:"editable"`
}

type GrafanaDatasources struct {
	apiVersion  string              `yaml:"apiVersion"`
	datasources []GrafanaDatasource `yaml:"datasources"`
}

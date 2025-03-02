package kubernetes

type KtIngressConfig struct {
	Target     string
	PrefixPath string
	Type       string
}

type KtTlsConfig struct {
	SecretName string
}

type KtHostConfig struct {
	Host      string
	TlsConfig *KtTlsConfig
	Ingresses []KtIngressConfig
}

type KtIngressClassConfig struct {
	//KtIngressClassName string
	//KtIngressConfigs []KtIngressConfig
	KtHostConfig []KtHostConfig
}

// KtMultiClassConfig mapping of Ingress class names onto config for that ingress
type KtMultiHostConfig map[string]KtHostConfig
type KtMultiClassConfig map[string]KtIngressClassConfig
type KtExtraIngressConfig struct {
	MultiIngressClassConfigs *KtMultiClassConfig
}

package kubernetes

type KtTlsConfig struct {
	SecretName string
}

type KtPortConfig struct {
	Name   *string
	Number *uint16
}

type KtHttpRule struct {
	PortConfig *KtPortConfig
	Path       string
	PathType   string
}

type KtAnnotations = map[string]string

type KtIngressSpec struct {
	IngressClassName *string
	Host             *string
	TlsConfig        *KtTlsConfig
	Annotations      *KtAnnotations
	HttpRules        []*KtHttpRule
}

type KtExtraIngressConfig struct {
	IngressSpecs []*KtIngressSpec
}

package golang

import "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang/generated"

func NewContextUuid(uuid string) *generated.ContextUuid {
	return &generated.ContextUuid{
		Value: uuid,
	}
}

func NewKurtosisContextConfig(currentContextUuid *generated.ContextUuid, contexts ...*generated.KurtosisContext) *generated.KurtosisContextsConfig {
	return &generated.KurtosisContextsConfig{
		CurrentContextUuid: currentContextUuid,
		Contexts:           contexts,
	}
}

func NewLocalOnlyContext(uuid *generated.ContextUuid, name string) *generated.KurtosisContext {
	return &generated.KurtosisContext{
		Uuid: uuid,
		Name: name,
		KurtosisContextInfo: &generated.KurtosisContext_LocalOnlyContextV0{
			LocalOnlyContextV0: &generated.LocalOnlyContextV0{},
		},
	}
}

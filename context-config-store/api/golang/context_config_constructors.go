package golang

func NewContextUuid(uuid string) *ContextUuid {
	return &ContextUuid{
		Value: uuid,
	}
}

func NewKurtosisContextConfig(currentContextUuid *ContextUuid, contexts ...*KurtosisContext) *KurtosisContextConfig {
	return &KurtosisContextConfig{
		CurrentContext: currentContextUuid,
		Contexts:       contexts,
	}
}

func NewLocalOnlyContext(uuid *ContextUuid, name string) *KurtosisContext {
	return &KurtosisContext{
		Uuid: uuid,
		Name: name,
		KurtosisContextInfo: &KurtosisContext_LocalOnlyContextV0{
			LocalOnlyContextV0: &LocalOnlyContextV0{},
		},
	}
}

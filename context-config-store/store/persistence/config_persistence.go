package persistence

import (
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"
)

type ConfigPersistence interface {
	// init initializes the persistence. It does nothing it's already initialized
	init() error

	// PersistContextConfig writes the context config to the persistence
	PersistContextConfig(contextConfig *api.KurtosisContextConfig) error

	// LoadContextConfig reads the context config from the persistence
	LoadContextConfig() (*api.KurtosisContextConfig, error)
}

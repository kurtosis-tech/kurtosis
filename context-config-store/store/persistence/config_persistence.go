package persistence

import (
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang/generated"
)

type ConfigPersistence interface {
	// init initializes the persistence. It does nothing if it is already initialized
	init() error

	// PersistContextConfig writes the context config to the persistence
	PersistContextConfig(contextConfig *api.KurtosisContextsConfig) error

	// LoadContextConfig reads the context config from the persistence
	LoadContextConfig() (*api.KurtosisContextsConfig, error)
}

package persistence

import (
	api "github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
)

type ConfigPersistence interface {
	// init initializes the persistence. It does nothing if it is already initialized
	init() error

	// PersistContextsConfig writes the context config to the persistence
	PersistContextsConfig(newContextsConfig *api.KurtosisContextsConfig) error

	// LoadContextsConfig reads the context config from the persistence
	LoadContextsConfig() (*api.KurtosisContextsConfig, error)
}

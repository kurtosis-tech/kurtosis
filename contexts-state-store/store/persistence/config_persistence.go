package persistence

import (
	api "github.com/kurtosis-tech/kurtosis/contexts-state-store/api/golang/generated"
)

type ConfigPersistence interface {
	// init initializes the persistence. It does nothing if it is already initialized
	init() error

	// PersistContextsState writes the context config to the persistence
	PersistContextsState(newContextsState *api.KurtosisContextsState) error

	// LoadContextsState reads the context config from the persistence
	LoadContextsState() (*api.KurtosisContextsState, error)
}

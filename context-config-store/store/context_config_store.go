package store

import (
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/context-config-store/store/persistence"
	"sync"
)

var (
	once               sync.Once
	contextConfigStore ContextConfigStore
)

type ContextConfigStore interface {
	// GetAllContexts returns the list of currently saved contexts.
	GetAllContexts() ([]*api.KurtosisContext, error)

	// GetCurrentContext returns the current context information.
	GetCurrentContext() (*api.KurtosisContext, error)

	// SwitchContext switches to the context passed as an argument.
	// It throws an error is the contextUuid does not.
	SwitchContext(contextUuid *api.ContextUuid) error

	// AddNewContext adds a new context to the store.
	// If the context already exists, it does nothing. Note that context equality is computed on
	// context UUID, nothing else. To update a context, consider removing it and adding it again.
	AddNewContext(contextToAdd *api.KurtosisContext) error

	// RemoveContext removes the contexts passed as an argument.
	// It does nothing is the contextUuid does not point to any known context.
	RemoveContext(contextUuid *api.ContextUuid) error
}

func GetContextConfigStore() ContextConfigStore {
	once.Do(func() {
		contextConfigStore = NewContextConfigStore(persistence.NewFileBackedConfigPersistence())
	})
	return contextConfigStore
}

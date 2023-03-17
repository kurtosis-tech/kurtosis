package store

import (
	"github.com/kurtosis-tech/kurtosis/context-config-store/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/context-config-store/store/persistence"
	"sync"
)

var (
	once               sync.Once
	contextConfigStore ContextConfigStore
)

type ContextConfigStore interface {
	// GetAllContexts returns the list of currently saved contexts.
	GetAllContexts() ([]*generated.KurtosisContext, error)

	// GetCurrentContext returns the current context information.
	GetCurrentContext() (*generated.KurtosisContext, error)

	// SwitchContext switches to the context passed as an argument.
	// It throws an error if the contextUuid does not point to any known context.
	SwitchContext(contextUuid *generated.ContextUuid) error

	// AddNewContext adds a new context to the store.
	// It throws an error if a context with the same UUID already exists
	AddNewContext(contextToAdd *generated.KurtosisContext) error

	// RemoveContext removes the contexts passed as an argument.
	// It does nothing if the contextUuid does not point to any known context.
	RemoveContext(contextUuid *generated.ContextUuid) error
}

func GetContextConfigStore() ContextConfigStore {
	once.Do(func() {
		contextConfigStore = NewContextConfigStore(persistence.NewFileBackedConfigPersistence())
	})
	return contextConfigStore
}

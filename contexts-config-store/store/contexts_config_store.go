package store

import (
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store/persistence"
	"sync"
)

var (
	once                sync.Once
	contextsConfigStore ContextsConfigStore
)

type ContextsConfigStore interface {
	// GetKurtosisContextsConfig returns the currently saved contexts configuration.
	GetKurtosisContextsConfig() (*generated.KurtosisContextsConfig, error)

	// GetCurrentContext returns the current context information.
	GetCurrentContext() (*generated.KurtosisContext, error)

	// SetContext switches to the context passed as an argument.
	// It throws an error if the contextUuid does not point to any known context.
	SetContext(contextUuid *generated.ContextUuid) error

	// AddNewContext adds a new context to the store.
	// It throws an error if a context with the same UUID already exists
	AddNewContext(newContext *generated.KurtosisContext) error

	// RemoveContext removes the contexts passed as an argument.
	// It does nothing if the contextUuid does not point to any known context.
	RemoveContext(contextUuid *generated.ContextUuid) error
}

func GetContextsConfigStore() ContextsConfigStore {
	once.Do(func() {
		contextsConfigStore = NewContextConfigStore(persistence.NewFileBackedConfigPersistence())
	})
	return contextsConfigStore
}

func IsRemote(kurtosisContext *generated.KurtosisContext) bool {
	var isRemote bool
	_, _ = golang.Visit[struct{}](kurtosisContext, golang.KurtosisContextVisitor[struct{}]{
		VisitLocalOnlyContextV0: func(_ *generated.LocalOnlyContextV0) (*struct{}, error) {
			isRemote = false
			return nil, nil
		},
		VisitRemoteContextV0: func(_ *generated.RemoteContextV0) (*struct{}, error) {
			isRemote = true
			return nil, nil
		},
	})
	return isRemote
}

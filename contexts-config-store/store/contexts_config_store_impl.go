package store

import (
	api "github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store/persistence"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
	"sync"
)

const (
	contextUuidsSeparator = ", "
)

type contextConfigStoreImpl struct {
	*sync.RWMutex

	storage persistence.ConfigPersistence
}

func NewContextConfigStore(storage persistence.ConfigPersistence) ContextsConfigStore {
	return &contextConfigStoreImpl{
		RWMutex: &sync.RWMutex{},
		storage: storage,
	}
}

func (store *contextConfigStoreImpl) GetKurtosisContextsConfig() (*generated.KurtosisContextsConfig, error) {
	store.RLock()
	defer store.RUnlock()

	contextsConfig, err := store.storage.LoadContextsConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to load the list of contexts currently stored")
	}
	return contextsConfig, nil
}

func (store *contextConfigStoreImpl) GetCurrentContext() (*generated.KurtosisContext, error) {
	store.RLock()
	defer store.RUnlock()

	contextsConfig, err := store.storage.LoadContextsConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to load the list of contexts currently stored")
	}

	currentContextUuid := contextsConfig.GetCurrentContextUuid()
	var contextUuidsInStore []string
	for _, kurtosisContextInStore := range contextsConfig.GetContexts() {
		contextUuidsInStore = append(contextUuidsInStore, kurtosisContextInStore.GetUuid().GetValue())
		if kurtosisContextInStore.GetUuid().GetValue() == currentContextUuid.GetValue() {
			return kurtosisContextInStore, nil
		}
	}
	return nil, stacktrace.NewError("Unable to find current context info in currently stored contexts config. "+
		"Current context is set to '%s' but known contexts are: '%s'",
		currentContextUuid.GetValue(),
		strings.Join(contextUuidsInStore, contextUuidsSeparator))
}

func (store *contextConfigStoreImpl) SetContext(contextUuid *generated.ContextUuid) error {
	store.Lock()
	defer store.Unlock()

	contextsConfig, err := store.storage.LoadContextsConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load the list of contexts currently stored")
	}

	var requestedContextDoesExist bool
	var contextUuidsInStore []string
	for _, kurtosisContextInStore := range contextsConfig.GetContexts() {
		contextUuidsInStore = append(contextUuidsInStore, kurtosisContextInStore.GetUuid().GetValue())
		if kurtosisContextInStore.GetUuid().GetValue() == contextUuid.Value {
			requestedContextDoesExist = true
		}
	}
	if !requestedContextDoesExist {
		return stacktrace.NewError("Context with UUID '%s' does not exist in store. Known contexts are: '%s'",
			contextUuid.GetValue(), strings.Join(contextUuidsInStore, contextUuidsSeparator))
	}

	newContextsConfigToPersist := api.NewKurtosisContextsConfig(contextUuid, contextsConfig.GetContexts()...)
	if err = store.storage.PersistContextsConfig(newContextsConfigToPersist); err != nil {
		return stacktrace.Propagate(err, "Unable to update current context in the context store")
	}
	return nil
}

func (store *contextConfigStoreImpl) AddNewContext(newContext *generated.KurtosisContext) error {
	store.Lock()
	defer store.Unlock()

	if newContext.GetName() == persistence.DefaultContextName {
		return stacktrace.NewError("Adding a new context with name '%s' is not allowed as it is a reserved "+
			"context name", persistence.DefaultContextName)
	}

	contextsConfig, err := store.storage.LoadContextsConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load the list of contexts currently stored")
	}

	newContextUuid := newContext.GetUuid()
	for _, kurtosisContextInStore := range contextsConfig.GetContexts() {
		if kurtosisContextInStore.GetUuid().GetValue() == newContextUuid.GetValue() {
			return stacktrace.NewError("Trying to add a context with UUID '%s' but a context already exist with "+
				"this UUID and name: '%s'. If the context should be replaced or updated, it should be removed first",
				newContextUuid.GetValue(), kurtosisContextInStore.GetName())
		}
	}

	var updatedContextsList []*generated.KurtosisContext
	updatedContextsList = append(updatedContextsList, contextsConfig.GetContexts()...)
	updatedContextsList = append(updatedContextsList, newContext)

	newContextConfigToPersist := api.NewKurtosisContextsConfig(
		contextsConfig.GetCurrentContextUuid(), updatedContextsList...)
	if err = store.storage.PersistContextsConfig(newContextConfigToPersist); err != nil {
		return stacktrace.Propagate(err, "Unable to persist new context config to store")
	}
	return nil
}

func (store *contextConfigStoreImpl) RemoveContext(contextUuid *generated.ContextUuid) error {
	store.Lock()
	defer store.Unlock()

	contextsConfig, err := store.storage.LoadContextsConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load the list of contexts currently stored")
	}

	if contextUuid.GetValue() == contextsConfig.GetCurrentContextUuid().GetValue() {
		return stacktrace.NewError("Cannot remove context '%s' as it is currently the selected context. "+
			"Switch to a different context before removing it", contextUuid.GetValue())
	}

	foundContextToRemove := false
	var newContexts []*generated.KurtosisContext
	for _, kurtosisContextInStore := range contextsConfig.GetContexts() {
		if kurtosisContextInStore.GetUuid().GetValue() == contextUuid.GetValue() {
			if kurtosisContextInStore.GetName() == persistence.DefaultContextName {
				return stacktrace.NewError("Cannot remove context '%s' as it is a reserved context",
					persistence.DefaultContextName)
			}
			foundContextToRemove = true
		} else {
			newContexts = append(newContexts, kurtosisContextInStore)
		}
	}

	if !foundContextToRemove {
		return nil
	}

	newContextConfigToPersist := api.NewKurtosisContextsConfig(contextsConfig.GetCurrentContextUuid(), newContexts...)
	if err = store.storage.PersistContextsConfig(newContextConfigToPersist); err != nil {
		return stacktrace.Propagate(err, "Unable to persist new context config to store")
	}
	return nil
}

package store

import (
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/context-config-store/store/persistence"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
	"sync"
)

type contextConfigStoreImpl struct {
	*sync.RWMutex

	storage persistence.ConfigPersistence
}

func NewContextConfigStore(storage persistence.ConfigPersistence) ContextConfigStore {
	return &contextConfigStoreImpl{
		RWMutex: &sync.RWMutex{},
		storage: storage,
	}
}

func (store *contextConfigStoreImpl) GetAllContexts() ([]*api.KurtosisContext, error) {
	store.RLock()
	defer store.RUnlock()

	contextConfig, err := store.storage.LoadContextConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to load current context config")
	}
	return contextConfig.GetContexts(), nil
}

func (store *contextConfigStoreImpl) GetCurrentContext() (*api.KurtosisContext, error) {
	store.RLock()
	defer store.RUnlock()

	contextConfig, err := store.storage.LoadContextConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to load current context config")
	}

	currentContextUuid := contextConfig.GetCurrentContext()
	var knownContexts []string
	for _, kurtosisContext := range contextConfig.GetContexts() {
		knownContexts = append(knownContexts, kurtosisContext.GetUuid().GetValue())
		if kurtosisContext.GetUuid().GetValue() == currentContextUuid.GetValue() {
			return kurtosisContext, nil
		}
	}
	return nil, stacktrace.NewError("Unable to find current context info in context config file. "+
		"Current context is set to '%s' but known contexts are: '%s'",
		currentContextUuid.GetValue(),
		strings.Join(knownContexts, ", "))
}

func (store *contextConfigStoreImpl) SwitchContext(contextUuid *api.ContextUuid) error {
	store.Lock()
	defer store.Unlock()

	contextConfig, err := store.storage.LoadContextConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load current context config")
	}

	var doesExist bool
	for _, kurtosisContext := range contextConfig.GetContexts() {
		doesExist = doesExist || kurtosisContext.GetUuid().GetValue() == contextUuid.Value
	}
	if !doesExist {
		return stacktrace.NewError("Context with UUID '%s' does not exist yet.", contextUuid.GetValue())
	}

	newContextConfig := api.NewKurtosisContextConfig(contextUuid, contextConfig.GetContexts()...)
	if err = store.storage.PersistContextConfig(newContextConfig); err != nil {
		return stacktrace.Propagate(err, "Unable to persist new current context to store")
	}
	return nil
}

func (store *contextConfigStoreImpl) AddNewContext(contextToAdd *api.KurtosisContext) error {
	store.Lock()
	defer store.Unlock()

	contextConfig, err := store.storage.LoadContextConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load current context config")
	}

	newContextUuid := contextToAdd.GetUuid()
	for _, kurtosisContext := range contextConfig.GetContexts() {
		if kurtosisContext.GetUuid().GetValue() == newContextUuid.GetValue() {
			return nil // doing nothing as context already exists
		}
	}

	var newContexts []*api.KurtosisContext
	for _, kurtosisContext := range contextConfig.GetContexts() {
		newContexts = append(newContexts, kurtosisContext)
	}
	newContexts = append(newContexts, contextToAdd)

	newContextConfig := api.NewKurtosisContextConfig(contextConfig.GetCurrentContext(), newContexts...)
	if err = store.storage.PersistContextConfig(newContextConfig); err != nil {
		return stacktrace.Propagate(err, "Unable to persist new context config to store")
	}
	return nil
}

func (store *contextConfigStoreImpl) RemoveContext(contextUuid *api.ContextUuid) error {
	store.Lock()
	defer store.Unlock()

	contextConfig, err := store.storage.LoadContextConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to load current context config")
	}

	foundContextToRemove := false
	var newContexts []*api.KurtosisContext
	for _, kurtosisContext := range contextConfig.GetContexts() {
		if kurtosisContext.GetUuid().GetValue() != contextUuid.GetValue() {
			newContexts = append(newContexts, kurtosisContext)
		} else {
			foundContextToRemove = true
		}
	}

	if !foundContextToRemove {
		return nil
	}

	newContextConfig := api.NewKurtosisContextConfig(contextConfig.GetCurrentContext(), newContexts...)
	if err = store.storage.PersistContextConfig(newContextConfig); err != nil {
		return stacktrace.Propagate(err, "Unable to persist new context config to store")
	}
	return nil
}

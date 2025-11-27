package store

import (
	"fmt"
	api "github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store/persistence"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"reflect"
	"testing"
)

const (
	// Unfortunately Mockery does not fail when we call AssertNotCalled("MethodNameWithATypo")
	// So we have to manually validate the method exist in the mock first, using reflect.MethodByName
	persistMethodName = "PersistContextsConfig"
)

var (
	contextUuid      = api.NewContextUuid("context-uuid")
	otherContextUuid = api.NewContextUuid("other-context-uuid")

	localContext      = api.NewLocalOnlyContext(contextUuid, "context-name")
	otherLocalContext = api.NewLocalOnlyContext(otherContextUuid, "other-context-name")
)

func TestGetAllContexts(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	result, err := testContextConfigStore.GetKurtosisContextsConfig()
	require.NoError(t, err)
	require.True(t, proto.Equal(contextsConfig, result))
}

func TestGetCurrentContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	result, err := testContextConfigStore.GetCurrentContext()
	require.NoError(t, err)
	require.Equal(t, result, localContext)
}

func TestGetCurrentContext_failureInconsistentContextConfig(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(otherContextUuid, localContext) // unknown current context UUID
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	result, err := testContextConfigStore.GetCurrentContext()
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Unable to find current context info in currently stored contexts config. "+
		"Current context is set to '%s' but known contexts are: '%s'",
		otherContextUuid.GetValue(), contextUuid.GetValue())
	require.Contains(t, err.Error(), expectedErr)
	require.Nil(t, result)
}

func TestSetContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	expectContextConfigAfterSwitch := api.NewKurtosisContextsConfig(otherContextUuid, localContext, otherLocalContext)
	storage.EXPECT().PersistContextsConfig(expectContextConfigAfterSwitch).Times(1).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.SetContext(otherContextUuid)
	require.NoError(t, err)
}

func TestSetContext_NonExistingContextFailure(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.SetContext(otherContextUuid)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Context with UUID '%s' does not exist in store. Known contexts are: '%s'",
		otherContextUuid.GetValue(), contextUuid.GetValue())
	require.Contains(t, err.Error(), expectedErr)
}

func TestAddNewContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	expectContextConfigAfterAddition := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().PersistContextsConfig(expectContextConfigAfterAddition).Times(1).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.AddNewContext(otherLocalContext)
	require.NoError(t, err)
}

func TestAddNewContext_AlreadyExists(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.AddNewContext(otherLocalContext)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Trying to add a context with UUID '%s' but a context already exist with this "+
		"UUID and name: '%s'. If the context should be replaced or updated, it should be removed first",
		otherContextUuid.GetValue(), otherLocalContext.GetName())
	require.Contains(t, err.Error(), expectedErr)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	method, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, method.Name, mock.Anything)
}

func TestAddNewContext_DefaultContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	newDefaultContext := api.NewLocalOnlyContext(contextUuid, persistence.DefaultContextName)
	err := testContextConfigStore.AddNewContext(newDefaultContext)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Adding a new context with name '%s' is not allowed as it is a reserved context name",
		persistence.DefaultContextName)
	require.Contains(t, err.Error(), expectedErr)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	method, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, method.Name, mock.Anything)
}

func TestRemoveContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	expectContextsConfigAfterRemoval := api.NewKurtosisContextsConfig(contextUuid, localContext)
	storage.EXPECT().PersistContextsConfig(expectContextsConfigAfterRemoval).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(otherContextUuid)
	require.NoError(t, err)
}

func TestRemoveContext_FailureCurrentContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(contextUuid)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Cannot remove context '%s' as it is currently the selected context. Switch to a "+
		"different context before removing it", contextUuid.GetValue())
	require.Contains(t, err.Error(), expectedErr)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	persistMethod, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, persistMethod.Name, mock.Anything)
}

func TestRemoveContext_FailureDefaultContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	defaultContextUuid := api.NewContextUuid("default-context-uuid")
	defaultDefaultContext := api.NewLocalOnlyContext(defaultContextUuid, persistence.DefaultContextName)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, defaultDefaultContext, localContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(defaultContextUuid)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Cannot remove context '%s' as it is a reserved context",
		persistence.DefaultContextName)
	require.Contains(t, err.Error(), expectedErr)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	persistMethod, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, persistMethod.Name, mock.Anything)
}

func TestRemoveContext_NonExistingContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextsConfig := api.NewKurtosisContextsConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextsConfig().Return(contextsConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(otherContextUuid)
	require.NoError(t, err)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	persistMethod, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, persistMethod.Name, mock.Anything)
}

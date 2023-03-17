package store

import (
	"fmt"
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/context-config-store/store/persistence"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

const (
	// Unfortunately Mockery does not fail when we call AssertNotCalled("MethodNameWithATypo")
	// So we have to manually validate the method exist in the mock first, using reflect.MethodByName
	persistMethodName = "PersistContextConfig"
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
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	result, err := testContextConfigStore.GetAllContexts()
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Contains(t, result, localContext)
}

func TestGetCurrentContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	result, err := testContextConfigStore.GetCurrentContext()
	require.NoError(t, err)
	require.Equal(t, result, localContext)
}

func TestGetCurrentContext_failureInconsistentContextConfig(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(otherContextUuid, localContext) // unknown current context UUID
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

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

func TestSwitchContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	expectContextConfigAfterSwitch := api.NewKurtosisContextConfig(otherContextUuid, localContext, otherLocalContext)
	storage.EXPECT().PersistContextConfig(expectContextConfigAfterSwitch).Times(1).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.SwitchContext(otherContextUuid)
	require.NoError(t, err)
}

func TestSwitchContext_NonExistingContextFailure(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.SwitchContext(otherContextUuid)
	require.Error(t, err)
	expectedErr := fmt.Sprintf("Context with UUID '%s' does not exist in store. Known contexts are: '%s'",
		otherContextUuid.GetValue(), contextUuid.GetValue())
	require.Contains(t, err.Error(), expectedErr)
}

func TestAddNewContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	expectContextConfigAfterAddition := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().PersistContextConfig(expectContextConfigAfterAddition).Times(1).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.AddNewContext(otherLocalContext)
	require.NoError(t, err)
}

func TestAddNewContext_AlreadyExists(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

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

func TestRemoveContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	expectContextConfigAfterRemoval := api.NewKurtosisContextConfig(contextUuid, localContext)
	storage.EXPECT().PersistContextConfig(expectContextConfigAfterRemoval).Return(nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(otherContextUuid)
	require.NoError(t, err)
}

func TestRemoveContext_FailureCurrentContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext, otherLocalContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(contextUuid)
	require.Error(t, err)
	expecteErr := fmt.Sprintf("Cannot remove context '%s' as it is currently the selected context. Switch to a "+
		"different context before removing it", contextUuid.GetValue())
	require.Contains(t, err.Error(), expecteErr)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	persistMethod, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, persistMethod.Name, mock.Anything)
}

func TestRemoveContext_NonExistingContext(t *testing.T) {
	// Setup storage mock
	storage := persistence.NewMockConfigPersistence(t)
	contextConfig := api.NewKurtosisContextConfig(contextUuid, localContext)
	storage.EXPECT().LoadContextConfig().Return(contextConfig, nil)

	// Run test
	testContextConfigStore := NewContextConfigStore(storage)
	err := testContextConfigStore.RemoveContext(otherContextUuid)
	require.NoError(t, err)

	// Need to check the method exist first because if the method name changes in the future this test would do nothing
	persistMethod, found := reflect.TypeOf(storage).MethodByName(persistMethodName)
	require.True(t, found)
	storage.AssertNotCalled(t, persistMethod.Name, mock.Anything)
}

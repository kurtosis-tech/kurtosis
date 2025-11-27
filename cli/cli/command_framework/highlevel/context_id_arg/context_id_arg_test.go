package context_id_arg

import (
	"fmt"
	store_api "github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	defaultContextUuid   = store_api.NewContextUuid("00000000000000000000000000000000")
	defaultShortenedUuid = "000000000000"
	defaultContextName   = "default"

	other1ContextUuid   = store_api.NewContextUuid("11111111111111111111111111111111")
	other1ShortenedUuid = "111111111111"
	other1ContextName   = "my-context"

	other2ContextUuid = store_api.NewContextUuid("22222222222222222222222222222222")
	other2ContextName = "my-other-context"
)

func TestGetContextUuidForContextIdentifier_WithShortenedUuids(t *testing.T) {
	mockContextsConfigStore := store.NewMockContextsConfigStore(t)
	contextsConfig := store_api.NewKurtosisContextsConfig(
		defaultContextUuid,
		store_api.NewLocalOnlyContext(defaultContextUuid, defaultContextName),
		store_api.NewLocalOnlyContext(other1ContextUuid, other1ContextName),
		store_api.NewLocalOnlyContext(other2ContextUuid, other2ContextName),
	)
	mockContextsConfigStore.EXPECT().GetKurtosisContextsConfig().Return(
		contextsConfig,
		nil,
	)

	requestedContextIdentifiers := []string{
		other1ShortenedUuid,
		defaultShortenedUuid,
	}
	result, err := GetContextUuidForContextIdentifier(mockContextsConfigStore, requestedContextIdentifiers)
	require.NoError(t, err)

	require.Len(t, result, 2)
	require.Equal(t, defaultContextUuid, result[defaultShortenedUuid])
	require.Equal(t, other1ContextUuid, result[other1ShortenedUuid])
}

func TestGetContextUuidForContextIdentifier_WithFullUuids(t *testing.T) {
	mockContextsConfigStore := store.NewMockContextsConfigStore(t)
	contextsConfig := store_api.NewKurtosisContextsConfig(
		defaultContextUuid,
		store_api.NewLocalOnlyContext(defaultContextUuid, defaultContextName),
		store_api.NewLocalOnlyContext(other1ContextUuid, other1ContextName),
		store_api.NewLocalOnlyContext(other2ContextUuid, other2ContextName),
	)
	mockContextsConfigStore.EXPECT().GetKurtosisContextsConfig().Return(
		contextsConfig,
		nil,
	)

	requestedContextIdentifiers := []string{
		other1ContextUuid.GetValue(),
		defaultContextUuid.GetValue(),
	}
	result, err := GetContextUuidForContextIdentifier(mockContextsConfigStore, requestedContextIdentifiers)
	require.NoError(t, err)

	require.Len(t, result, 2)
	require.Equal(t, defaultContextUuid, result[defaultContextUuid.GetValue()])
	require.Equal(t, other1ContextUuid, result[other1ContextUuid.GetValue()])
}

func TestGetContextUuidForContextIdentifier_WithNames(t *testing.T) {
	mockContextsConfigStore := store.NewMockContextsConfigStore(t)
	contextsConfig := store_api.NewKurtosisContextsConfig(
		defaultContextUuid,
		store_api.NewLocalOnlyContext(defaultContextUuid, defaultContextName),
		store_api.NewLocalOnlyContext(other1ContextUuid, other1ContextName),
		store_api.NewLocalOnlyContext(other2ContextUuid, other2ContextName),
	)
	mockContextsConfigStore.EXPECT().GetKurtosisContextsConfig().Return(
		contextsConfig,
		nil,
	)

	requestedContextIdentifiers := []string{
		other1ContextName,
		defaultContextName,
	}
	result, err := GetContextUuidForContextIdentifier(mockContextsConfigStore, requestedContextIdentifiers)
	require.NoError(t, err)

	require.Len(t, result, 2)
	require.Equal(t, defaultContextUuid, result[defaultContextName])
	require.Equal(t, other1ContextUuid, result[other1ContextName])
}

func TestGetContextUuidForContextIdentifier_UnknownIdentifier(t *testing.T) {
	mockContextsConfigStore := store.NewMockContextsConfigStore(t)
	contextsConfig := store_api.NewKurtosisContextsConfig(
		defaultContextUuid,
		store_api.NewLocalOnlyContext(defaultContextUuid, defaultContextName),
		store_api.NewLocalOnlyContext(other1ContextUuid, other1ContextName),
		store_api.NewLocalOnlyContext(other2ContextUuid, other2ContextName),
	)
	mockContextsConfigStore.EXPECT().GetKurtosisContextsConfig().Return(
		contextsConfig,
		nil,
	)

	unknownIdentifier := "unknownidentifier"
	requestedContextIdentifiers := []string{
		unknownIdentifier,
		defaultContextName,
	}
	result, err := GetContextUuidForContextIdentifier(mockContextsConfigStore, requestedContextIdentifiers)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("No context found for identifier '%s'", unknownIdentifier))
	require.Nil(t, result)
}

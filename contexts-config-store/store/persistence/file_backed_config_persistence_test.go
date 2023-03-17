package persistence

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

const (
	tempFileDir         = ""
	tempFileNamePattern = "context-config-persistence-testing-*.json"

	contextUuid = "context-uuid"
	contextName = "context-name"
)

var (
	contextConfig = &generated.KurtosisContextsConfig{
		CurrentContextUuid: &generated.ContextUuid{Value: contextUuid},
		Contexts: []*generated.KurtosisContext{
			{
				Uuid: &generated.ContextUuid{Value: contextUuid},
				Name: contextName,
				KurtosisContextInfo: &generated.KurtosisContext_LocalOnlyContextV0{
					LocalOnlyContextV0: &generated.LocalOnlyContextV0{},
				},
			},
		},
	}

	serializedContextConfig = "{\"currentContextUuid\":{\"value\":\"context-uuid\"},\"contexts\":[{\"uuid\":{\"value\":\"context-uuid\"},\"name\":\"context-name\",\"localOnlyContextV0\":{}}]}"
)

func TestPersistConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)
	defer os.Remove(tempFile.Name())

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	err = storage.PersistContextsConfig(contextConfig)
	require.Nil(t, err)

	fileContent, err := os.ReadFile(tempFile.Name())
	require.Nil(t, err)
	require.Equal(t, serializedContextConfig, string(fileContent))
}

func TestLoadConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(serializedContextConfig))
	require.Nil(t, err)

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	result, err := storage.LoadContextsConfig()
	require.Nil(t, err)
	require.True(t, proto.Equal(contextConfig, result))
}

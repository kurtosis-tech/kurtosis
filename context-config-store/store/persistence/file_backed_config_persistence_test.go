package persistence

import (
	api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"
	"github.com/stretchr/testify/require"
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
	contextConfig = &api.KurtosisContextConfig{
		CurrentContext: &api.ContextUuid{Value: contextUuid},
		Contexts: []*api.KurtosisContext{
			{
				Uuid:                &api.ContextUuid{Value: contextUuid},
				Name:                contextName,
				KurtosisContextInfo: &api.KurtosisContext_LocalOnlyContextV0{},
			},
		},
	}

	serializedContextConfig = `{
  "currentContext": {
    "value": "context-uuid"
  },
  "contexts": [
    {
      "uuid": {
        "value": "context-uuid"
      },
      "name": "context-name",
      "localOnlyContextV0": {}
    }
  ]
}`
)

func TestPersistConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	err = storage.PersistContextConfig(contextConfig)
	require.Nil(t, err)

	fileContent, err := os.ReadFile(tempFile.Name())
	require.Nil(t, err)
	require.Equal(t, serializedContextConfig, string(fileContent))
}

func TestLoadConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)
	_, err = tempFile.Write([]byte(serializedContextConfig))
	require.Nil(t, err)

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	result, err := storage.LoadContextConfig()
	require.Nil(t, err)
	require.Equal(t, contextConfig.GetCurrentContext().GetValue(), result.GetCurrentContext().GetValue())
	require.Len(t, result.GetContexts(), len(contextConfig.GetContexts()))
	require.Equal(t, contextConfig.GetContexts()[0].GetUuid().GetValue(), result.GetContexts()[0].GetUuid().GetValue())
}

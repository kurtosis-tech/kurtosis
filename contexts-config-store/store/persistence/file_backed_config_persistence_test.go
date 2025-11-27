package persistence

import (
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
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

	serializedContextsConfig = "{\"contexts\":[{\"localOnlyContextV0\":{},\"name\":\"context-name\",\"uuid\":{\"value\":\"context-uuid\"}}],\"currentContextUuid\":{\"value\":\"context-uuid\"}}"
)

func TestPersistConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)
	defer os.Remove(tempFile.Name())

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	err = storage.PersistContextsConfig(contextConfig)
	require.Nil(t, err)

	fileContent, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	// protojson provides no guarantee for the shape of the json produced, they just provide compatibility between
	// protojson.Unmarshal and protojson.Marshal. To be able to validate output with the expected JSON, we need to
	// compare the actual proto messages. See:
	// - https://protobuf.dev/reference/go/faq/#unstable-json
	// - https://github.com/golang/protobuf/issues/1082
	writtenContextsConfig := new(generated.KurtosisContextsConfig)
	require.NoError(t, protojson.Unmarshal(fileContent, writtenContextsConfig))
	expectedContextsConfig := new(generated.KurtosisContextsConfig)
	require.NoError(t, protojson.Unmarshal([]byte(serializedContextsConfig), expectedContextsConfig))

	require.True(t, proto.Equal(writtenContextsConfig, expectedContextsConfig))
}

func TestLoadConfig(t *testing.T) {
	tempFile, err := os.CreateTemp(tempFileDir, tempFileNamePattern)
	require.Nil(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(serializedContextsConfig))
	require.Nil(t, err)

	storage := newFileBackedConfigPersistenceForTesting(tempFile.Name())
	result, err := storage.LoadContextsConfig()
	require.Nil(t, err)
	require.True(t, proto.Equal(contextConfig, result))
}

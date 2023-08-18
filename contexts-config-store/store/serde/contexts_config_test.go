package serde

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"testing"
)

const (
	defaultContextUuid = "dc2a9471c70649b7aafae2448970dd2e"
	defaultContextName = "default"

	otherContextUuid = "54702bfbc6f242bea540134a77186d22"
	otherContextName = "other"
)

var (
	contextConfig = &generated.KurtosisContextsConfig{
		CurrentContextUuid: &generated.ContextUuid{Value: defaultContextUuid},
		Contexts: []*generated.KurtosisContext{
			{
				Uuid: &generated.ContextUuid{Value: defaultContextUuid},
				Name: defaultContextName,
				KurtosisContextInfo: &generated.KurtosisContext_LocalOnlyContextV0{
					LocalOnlyContextV0: &generated.LocalOnlyContextV0{},
				},
			},
			{
				Uuid: &generated.ContextUuid{Value: otherContextUuid},
				Name: otherContextName,
				KurtosisContextInfo: &generated.KurtosisContext_LocalOnlyContextV0{
					LocalOnlyContextV0: &generated.LocalOnlyContextV0{},
				},
			},
		},
	}

	serializedContextsConfig = `{
	"contexts":[{
		"localOnlyContextV0":{},
		"name":"default",
		"uuid":{
			"value":"dc2a9471c70649b7aafae2448970dd2e"
		}
	}, {
		"localOnlyContextV0":{},
		"name":"other",
		"uuid":{
			"value":"54702bfbc6f242bea540134a77186d22"
		}
	}],
	"currentContextUuid":{
		"value":"dc2a9471c70649b7aafae2448970dd2e"
	},
    "unknownField": "shouldBeIgnored"
}`
)

func TestSerializeContextsConfig(t *testing.T) {
	result, err := SerializeKurtosisContextsConfig(contextConfig)
	require.NoError(t, err)

	// protojson provides no guarantee for the shape of the json produced, they just provide compatibility between
	// protojson.Unmarshal and protojson.Marshal. To be able to validate output with the expected JSON, we need to
	// compare the actual proto messages. See:
	// - https://protobuf.dev/reference/go/faq/#unstable-json
	// - https://github.com/golang/protobuf/issues/1082
	writtenContextsConfig := new(generated.KurtosisContextsConfig)
	require.NoError(t, protojson.Unmarshal(result, writtenContextsConfig))
	expectedContextsConfig := new(generated.KurtosisContextsConfig)

	unmarshaller := protojson.UnmarshalOptions{DiscardUnknown: true} // nolint: exhaustruct
	require.NoError(t, unmarshaller.Unmarshal([]byte(serializedContextsConfig), expectedContextsConfig))

	require.True(t, proto.Equal(writtenContextsConfig, expectedContextsConfig))
}

func TestDeserializeContextsConfig(t *testing.T) {
	result, err := DeserializeKurtosisContextsConfig([]byte(serializedContextsConfig))
	require.NoError(t, err)

	require.Nil(t, err)
	require.True(t, proto.Equal(contextConfig, result))
}

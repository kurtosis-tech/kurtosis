package serde

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"testing"
)

const (
	remoteContextUuid = "54702bfbc6f242bea540134a77186d22"
	remoteContextName = "kurtosis-cloud"

	host        = "cloud.kurtosis.com"
	portalPort  = 9730
	backendPort = 9732
	tunnelPort  = 9731

	fakeCa   = "fake-ca"
	fakeCert = "fake-cert"
	fakeKey  = "fake-key"
)

var (
	kurtosisContext = &generated.KurtosisContext{
		Uuid: &generated.ContextUuid{Value: remoteContextUuid},
		Name: remoteContextName,
		KurtosisContextInfo: &generated.KurtosisContext_RemoteContextV0{
			RemoteContextV0: &generated.RemoteContextV0{
				Host:                host,
				RemotePortalPort:    portalPort,
				KurtosisBackendPort: backendPort,
				TunnelPort:          tunnelPort,
				TlsConfig: &generated.TlsConfig{
					CertificateAuthority: []byte(fakeCa),
					ClientCertificate:    []byte(fakeCert),
					ClientKey:            []byte(fakeKey),
				},
				EnvVars:         new(string),
				CloudInstanceId: new(string),
				CloudUserId:     new(string),
			},
		},
	}

	serializedKurtosisContext = `{
	"uuid":{
		"value":"54702bfbc6f242bea540134a77186d22"
	}, 
	"name":"kurtosis-cloud", 
	"remoteContextV0":{
		"host":"cloud.kurtosis.com", 
		"remotePortalPort":9730, 
		"kurtosisBackendPort":9732, 
		"tunnelPort":9731, 
		"tlsConfig":{
			"certificateAuthority":"ZmFrZS1jYQ==",
			"clientCertificate":"ZmFrZS1jZXJ0", 
			"clientKey":"ZmFrZS1rZXk="
		},
		"envVars": "",
		"cloudInstanceId": "",
		"cloudUserId": ""
	}
}`
)

func TestSerializeKurtosisContext(t *testing.T) {
	result, err := SerializeKurtosisContext(kurtosisContext)
	require.NoError(t, err)

	// protojson provides no guarantee for the shape of the json produced, they just provide compatibility between
	// protojson.Unmarshal and protojson.Marshal. To be able to validate output with the expected JSON, we need to
	// compare the actual proto messages. See:
	// - https://protobuf.dev/reference/go/faq/#unstable-json
	// - https://github.com/golang/protobuf/issues/1082
	writtenKurtosisContext := new(generated.KurtosisContext)
	require.NoError(t, protojson.Unmarshal(result, writtenKurtosisContext))
	expectedKurtosisContext := new(generated.KurtosisContext)
	require.NoError(t, protojson.Unmarshal([]byte(serializedKurtosisContext), expectedKurtosisContext))

	require.True(t, proto.Equal(writtenKurtosisContext, expectedKurtosisContext))
}

func TestDeserializeKurtosisContext(t *testing.T) {
	result, err := DeserializeKurtosisContext([]byte(serializedKurtosisContext))
	require.NoError(t, err)

	require.Nil(t, err)
	require.True(t, proto.Equal(kurtosisContext, result))
}

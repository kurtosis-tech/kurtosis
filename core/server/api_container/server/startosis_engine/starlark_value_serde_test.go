package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	port_spec2 "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestStarlarkValueSerde_Integer(t *testing.T) {
	val := starlark.MakeInt(42)

	serializedStarlarkValue := SerializeStarlarkValue(val)
	require.Equal(t, "42", serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := DeserializeStarlarkValue(serializedStarlarkValue)
	require.Nil(t, interpretationErr)
	require.Equal(t, val, deserializedStarlarkValue)
}

func TestStarlarkValueSerde_String(t *testing.T) {
	val := starlark.String("Hello world")

	serializedStarlarkValue := SerializeStarlarkValue(val)
	require.Equal(t, `"Hello world"`, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := DeserializeStarlarkValue(serializedStarlarkValue)
	require.Nil(t, interpretationErr)
	require.Equal(t, val, deserializedStarlarkValue)
}

func TestStarlarkValueSerde_Dict(t *testing.T) {
	val := starlark.NewDict(3)
	require.NoError(t, val.SetKey(starlark.String("hello"), starlark.String("world")))
	require.NoError(t, val.SetKey(starlark.String("answer"), starlark.MakeInt(42)))
	require.NoError(t, val.SetKey(starlark.String("nested"), starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"blah": starlark.String("blah"),
	})))

	serializedStarlarkValue := SerializeStarlarkValue(val)
	require.Equal(t, `{"hello": "world", "answer": 42, "nested": struct(blah = "blah")}`, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := DeserializeStarlarkValue(serializedStarlarkValue)
	require.Nil(t, interpretationErr)
	require.Equal(t, val, deserializedStarlarkValue)
}

func TestStarlarkValueSerde_Service(t *testing.T) {
	port, interpretationErr := port_spec2.CreatePortSpec(
		uint16(443),
		port_spec.TransportProtocol_TCP,
		nil,
		"10s",
	)
	require.Nil(t, interpretationErr)
	ports := starlark.NewDict(1)
	require.NoError(t, ports.SetKey(starlark.String("http"), port))

	serviceObj, interpretationErr := kurtosis_types.CreateService(
		"test-service",
		"test-service-hostname",
		"192.168.0.22",
		ports,
	)
	require.Nil(t, interpretationErr)

	serializedStarlarkValue := SerializeStarlarkValue(serviceObj)
	expectedSerializedServiceObj := `Service(name="test-service", hostname="test-service-hostname", ip_address="192.168.0.22", ports={"http": PortSpec(number=443, transport_protocol="TCP", wait="10s")})`
	require.Equal(t, expectedSerializedServiceObj, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := DeserializeStarlarkValue(serializedStarlarkValue)
	require.Nil(t, interpretationErr)
	require.Equal(t, serviceObj.String(), deserializedStarlarkValue.String())
}

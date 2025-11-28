package kurtosis_types

import (
	port_spec_core "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

type StarlarkValueSerdeSuite struct {
	suite.Suite

	serde *StarlarkValueSerde
}

func TestRunStarlarkValueSerdeSuite(t *testing.T) {
	suite.Run(t, new(StarlarkValueSerdeSuite))
}

func (suite *StarlarkValueSerdeSuite) SetupTest() {
	thread := &starlark.Thread{
		Name:       "test-serde-thread",
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make),

		ServiceTypeName:            starlark.NewBuiltin(ServiceTypeName, NewServiceType().CreateBuiltin()),
		port_spec.PortSpecTypeName: starlark.NewBuiltin(port_spec.PortSpecTypeName, port_spec.NewPortSpecType().CreateBuiltin()),
	}
	suite.serde = NewStarlarkValueSerde(thread, starlarkEnv)
}

func (suite *StarlarkValueSerdeSuite) TestStarlarkValueSerde_Integer() {
	val := starlark.MakeInt(42)

	serializedStarlarkValue := suite.serde.Serialize(val)
	require.Equal(suite.T(), "42", serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := suite.serde.Deserialize(serializedStarlarkValue)
	require.Nil(suite.T(), interpretationErr)
	require.Equal(suite.T(), val, deserializedStarlarkValue)
}

func (suite *StarlarkValueSerdeSuite) TestStarlarkValueSerde_String() {
	val := starlark.String("Hello world")

	serializedStarlarkValue := suite.serde.Serialize(val)
	require.Equal(suite.T(), `"Hello world"`, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := suite.serde.Deserialize(serializedStarlarkValue)
	require.Nil(suite.T(), interpretationErr)
	require.Equal(suite.T(), val, deserializedStarlarkValue)
}

func (suite *StarlarkValueSerdeSuite) TestStarlarkValueSerde_Dict() {
	val := starlark.NewDict(3)
	require.NoError(suite.T(), val.SetKey(starlark.String("hello"), starlark.String("world")))
	require.NoError(suite.T(), val.SetKey(starlark.String("answer"), starlark.MakeInt(42)))
	require.NoError(suite.T(), val.SetKey(starlark.String("nested"), starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"blah": starlark.String("blah"),
	})))

	serializedStarlarkValue := suite.serde.Serialize(val)
	require.Equal(suite.T(), `{"hello": "world", "answer": 42, "nested": struct(blah = "blah")}`, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := suite.serde.Deserialize(serializedStarlarkValue)
	require.Nil(suite.T(), interpretationErr)
	require.Equal(suite.T(), val, deserializedStarlarkValue)
}

func (suite *StarlarkValueSerdeSuite) TestStarlarkValueSerde_Service() {
	applicationProtocol := "http"
	maybeUrl := "http://my-test-service:443"
	port, interpretationErr := port_spec.CreatePortSpecUsingGoValues(
		"my-test-service",
		uint16(443),
		port_spec_core.TransportProtocol_TCP,
		&applicationProtocol,
		"10s",
		&maybeUrl,
	)
	require.Nil(suite.T(), interpretationErr)
	ports := starlark.NewDict(1)
	require.NoError(suite.T(), ports.SetKey(starlark.String("http"), port))

	serviceObj, interpretationErr := CreateService(
		"test-service",
		"test-service-hostname",
		"192.168.0.22",
		ports,
	)
	require.Nil(suite.T(), interpretationErr)

	serializedStarlarkValue := suite.serde.Serialize(serviceObj)
	expectedSerializedServiceObj := `Service(name="test-service", hostname="test-service-hostname", ip_address="192.168.0.22", ports={"http": PortSpec(number=443, transport_protocol="TCP", application_protocol="http", wait="10s", url="http://my-test-service:443")})`
	require.Equal(suite.T(), expectedSerializedServiceObj, serializedStarlarkValue)

	deserializedStarlarkValue, interpretationErr := suite.serde.Deserialize(serializedStarlarkValue)
	require.Nil(suite.T(), interpretationErr)
	require.Equal(suite.T(), serviceObj.String(), deserializedStarlarkValue.String())
}

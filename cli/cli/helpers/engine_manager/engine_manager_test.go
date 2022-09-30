package engine_manager

import (
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOneToOneMappingBetweenObjAttrProtosAndDockerProtos(t *testing.T) {
	// Ensure all obj attr protos are used
	require.Equal(t, len(schema.AllowedProtocols), len(objAttrsSchemaPortProtosToDockerPortProtos))

	// Ensure all teh declared obj attr protos are valid
	for candidateObjAttrProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		_, found := schema.AllowedProtocols[candidateObjAttrProto]
		require.True(t, found, "Invalid object attribute schema proto '%v'", candidateObjAttrProto)
	}

	// Ensure no duplicate Docker protos, which is the best we can do since Docker doesn't expose an enum of all the protos they support
	seenDockerProtos := map[string]schema.PortProtocol{}
	for objAttrProto, dockerProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		preexistingObjAttrProto, found := seenDockerProtos[dockerProto]
		require.False(t, found, "Docker proto '%v' is already in use by obj attr proto '%v'", dockerProto, preexistingObjAttrProto)
		seenDockerProtos[dockerProto] = objAttrProto
	}
}

package docker_compose_tranpsiler

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// From https://github.com/docker/awesome-compose/tree/master/minecraft
func TestMinecraftCompose(t *testing.T) {
	composeBytes := []byte(`
services:
 minecraft:
   image: itzg/minecraft-server
   ports:
     - "25565:25565"
   environment:
     EULA: "TRUE"
   deploy:
     resources:
       limits:
         memory: 1.5G
   volumes:
     - "~/minecraft_data:/data"
`)
	expectedResult := `def run(plan):
    plan.add_service(name = 'minecraft', config = ServiceConfig(image="itzg/minecraft-server", ports={"port0": PortSpec(number=25565, transport_protocol="TCP")}, env_vars={"EULA": "TRUE"}))
`

	result, err := convertComposeToStarlark(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

package docker_compose_transpiler

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
    plan.upload_files(src = "~/minecraft_data", name = "minecraft--volume0")
    plan.add_service(name = "minecraft", config = ServiceConfig(image="itzg/minecraft-server", ports={"port0": PortSpec(number=25565, transport_protocol="TCP")}, files={"/data": "minecraft--volume0"}, env_vars={"EULA": "TRUE"}))
`

	result, err := convertComposeToStarlark(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/tree/master/angular
func TestAngularCompose(t *testing.T) {
	composeBytes := []byte(`
 services:
  web:
    build:
      context: angular
      target: builder
    ports:
      - 4200:4200
    volumes:
      - ./angular:/project
      - /project/node_modules
`)
	expectedResult := `def run(plan):
    plan.upload_files(src = "./angular", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(context_dir="angular", target_stage="builder"), ports={"port0": PortSpec(number=4200, transport_protocol="TCP")}, files={"/project": "web--volume0", "/project/node_modules": ServiceConfig(persistent_key="volume1")}, env_vars={}))
`

	result, err := convertComposeToStarlark(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestAspnetComposeImageBuildSpec(t *testing.T) {
	composeBytes := []byte(`
services:
  web:
    build: app/aspnet
    ports:
      - 80:80
`)
	expectedResult := `def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(context_dir="app/aspnet"), ports={"port0": PortSpec(number=80, transport_protocol="TCP")}, env_vars={}))
`

	result, err := convertComposeToStarlark(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestDjangoComposeImageBuildSpecWithTarget(t *testing.T) {
	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports: 
      - '8000:8000'
`)
	expectedResult := `def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=8000, transport_protocol="TCP")}, env_vars={}))
`

	result, err := convertComposeToStarlark(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

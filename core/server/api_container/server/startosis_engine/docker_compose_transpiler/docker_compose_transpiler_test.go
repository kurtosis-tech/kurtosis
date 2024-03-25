package docker_compose_transpiler

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	testPackageAbsDirPathPattern = "package"
	testFilePerms                = 0644
)

// TODO: Create a test framework like starlark test framework so we updating tests manually with starlark or compose is less annoying

func TestMinimalCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web:
    image: app/server
    ports:
      - 80:80
`)

	expectedResult := `def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image="app/server", ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, env_vars={}))
`

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithImageBuildSpec(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web:
    build: app/server
    ports:
      - 80:80
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app/server"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithImageBuildSpecAndTarget(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports: 
      - 80:80
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithImageBuildSpecAndTargetAndName(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	// ignore the image name, and have kurtosis set it
	composeBytes := []byte(`
services:
  web:
    image: web
    build:
      context: app
      target: builder
    ports: 
      - 80:80
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithVolume(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	relVolumePath := "./data"
	err = os.Mkdir(path.Join(testPackageAbsDirPath, relVolumePath), testFilePerms)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports: 
      - '80:80'
    volumes:
     - ./data:/data
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.upload_files(src = "./data", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%v", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/data": "web--volume0"}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolume(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports: 
      - '80:80'
    volumes:
     - /project/node_modules
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/project/node_modules": Directory(persistent_key="web--volume0")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolumeAtProvidedPath(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports: 
      - '80:80'
    volumes:
     - /project/node_modules:/node_modules
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/node_modules": Directory(persistent_key="web--volume0")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolumeAtProvidedPathAreUnique(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web2: 
    build:
      context: app
      target: builder
    ports: 
      - '80:80'
    volumes:
     - /project/node_modules:/node_modules
  web3: 
    build:
      context: app
      target: builder
    ports: 
      - '80:80'
    volumes:
     - /project/node_modules:/node_modules
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web2", config = ServiceConfig(image=ImageBuildSpec(image_name="web2%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web2:80")}, files={"/node_modules": Directory(persistent_key="web2--volume0")}, env_vars={}))
    plan.add_service(name = "web3", config = ServiceConfig(image=ImageBuildSpec(image_name="web3%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web3:80")}, files={"/node_modules": Directory(persistent_key="web3--volume0")}, env_vars={}))
`, builtImageSuffix, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithEnvFile(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	relEnvFilePath := "./web.env"
	err = os.WriteFile(path.Join(testPackageAbsDirPath, relEnvFilePath), []byte("USERNAME=kurtosis"), testFilePerms)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    env_file:
      - ./web.env
    ports: 
      - 80:80
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, env_vars={"USERNAME": "kurtosis"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithNonRFC1035ServiceName(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  Web_Service: 
    build:
      context: app
      target: builder
    ports: 
      - 80:80
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "web-service", config = ServiceConfig(image=ImageBuildSpec(image_name="web-service%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web-service:80")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// Tests all supported compose functionalities for a single service
func TestFullCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	relVolumePath := "./data"
	err = os.Mkdir(path.Join(testPackageAbsDirPath, relVolumePath), testFilePerms)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  web: 
    build:
      context: app
      target: builder
    ports:
     - '80:80'
    environment:
     NODE_ENV: "development"
    volumes:
     - ./data:/data
     - /project/node_modules:/node_modules
    entrypoint:
     - /bin/echo
     - -c
     - echo "Hello"
    command: ["echo", "Hello,", "World!"]
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.upload_files(src = "./data", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%v", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/data": "web--volume0", "/node_modules": Directory(persistent_key="web--volume1")}, entrypoint=["/bin/echo", "-c", "echo \"Hello\""], cmd=["echo", "Hello,", "World!"], env_vars={"NODE_ENV": "development"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMultiServiceCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
 redis:
  image: 'redislabs/redismod'
  ports:
    - '6379:6379'
 web1:
  restart: on-failure
  build: ./web
  hostname: web1
  ports:
    - '81:5000'
 web2:
  restart: on-failure
  build: ./web
  hostname: web2
  ports:
    - '82:5000'
  depends_on:
  - redis
 nginx:
  build: ./nginx
  ports:
    - '80:80'
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "nginx", config = ServiceConfig(image=ImageBuildSpec(image_name="nginx%s", build_context_dir="./nginx"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://nginx:80")}, env_vars={}))
    plan.add_service(name = "redis", config = ServiceConfig(image="redislabs/redismod", ports={"port0": PortSpec(number=6379, transport_protocol="TCP")}, env_vars={}))
    plan.add_service(name = "web1", config = ServiceConfig(image=ImageBuildSpec(image_name="web1%s", build_context_dir="./web"), ports={"port0": PortSpec(number=5000, transport_protocol="TCP")}, env_vars={}))
    plan.add_service(name = "web2", config = ServiceConfig(image=ImageBuildSpec(image_name="web2%s", build_context_dir="./web"), ports={"port0": PortSpec(number=5000, transport_protocol="TCP")}, env_vars={}))
`, builtImageSuffix, builtImageSuffix, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// Simple tests for topological sort
func TestSortServiceBasedOnDependencies(t *testing.T) {
	perServiceDependencies := map[string]map[string]bool{
		"web":     {"nginx": true, "backend": true},
		"nginx":   {"backend": true},
		"backend": {},
	}

	expectedOrder := []string{"backend", "nginx", "web"}
	sortOrder, err := sortServicesBasedOnDependencies(perServiceDependencies)

	require.NoError(t, err)
	require.Equal(t, expectedOrder, sortOrder)
}

func TestSortServiceBasedOnDependenciesBreaksTiesDeterministically(t *testing.T) {
	perServiceDependencies := map[string]map[string]bool{
		"web":      {"nginx": true, "backend": true},
		"nginx":    {"backend": true},
		"backend":  {},
		"database": {},
	}

	// backend and database have no dependencies, but backend should come before because of lexicographic order
	expectedOrder := []string{"backend", "database", "nginx", "web"}
	sortOrder, err := sortServicesBasedOnDependencies(perServiceDependencies)

	require.NoError(t, err)
	require.Equal(t, expectedOrder, sortOrder)
}

func TestSortServiceBasedOnDependenciesWithCycle(t *testing.T) {
	perServiceDependencies := map[string]map[string]bool{
		"web":     {"nginx": true, "backend": true},
		"nginx":   {"backend": true},
		"backend": {"web": true},
	}

	_, err := sortServicesBasedOnDependencies(perServiceDependencies)
	require.Error(t, err)
}

func TestSortServiceBasedOnDependenciesWithNoDependencies(t *testing.T) {
	perServiceDependencies := map[string]map[string]bool{
		"web":     {},
		"nginx":   {},
		"backend": {},
	}

	// order should be alphabetical
	expectedOrder := []string{"backend", "nginx", "web"}
	actualOrder, err := sortServicesBasedOnDependencies(perServiceDependencies)

	require.NoError(t, err)
	require.Equal(t, expectedOrder, actualOrder)
}

func TestSortServiceBasedOnDependenciesWithLinearDependencies(t *testing.T) {
	perServiceDependencies := map[string]map[string]bool{
		"backend": {},
		"web":     {"backend": true},
		"nginx":   {"web": true},
	}

	expectedOrder := []string{"backend", "web", "nginx"}
	actualOrder, err := sortServicesBasedOnDependencies(perServiceDependencies)

	require.NoError(t, err)
	require.Equal(t, expectedOrder, actualOrder)
}

func TestMultiServiceComposeWithDependsOn(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
 redis:
  image: 'redislabs/redismod'
  ports:
    - '6379:6379'
 web1:
  restart: on-failure
  build: ./web
  hostname: web1
  ports:
    - '81:5000'
  depends_on:
  - redis
 web2:
  restart: on-failure
  build: ./web
  hostname: web2
  ports:
    - '82:5000'
  depends_on:
  - redis
 nginx:
  build: ./nginx
  ports:
    - '80:80'
  depends_on:
  - web1
  - web2
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "redis", config = ServiceConfig(image="redislabs/redismod", ports={"port0": PortSpec(number=6379, transport_protocol="TCP")}, env_vars={}))
    plan.add_service(name = "web1", config = ServiceConfig(image=ImageBuildSpec(image_name="web1%s", build_context_dir="./web"), ports={"port0": PortSpec(number=5000, transport_protocol="TCP")}, env_vars={}))
    plan.add_service(name = "web2", config = ServiceConfig(image=ImageBuildSpec(image_name="web2%s", build_context_dir="./web"), ports={"port0": PortSpec(number=5000, transport_protocol="TCP")}, env_vars={}))
    plan.add_service(name = "nginx", config = ServiceConfig(image=ImageBuildSpec(image_name="nginx%s", build_context_dir="./nginx"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://nginx:80")}, env_vars={}))
`, builtImageSuffix, builtImageSuffix, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// Test depends on with circular dependency returns error
func TestMultiServiceComposeWithCycleInDependsOn(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
 redis:
  image: 'redislabs/redismod'
  ports:
    - '6379:6379'
  depends_on:
  - nginx
 web1:
  restart: on-failure
  build: ./web
  hostname: web1
  ports:
    - '81:5000'
  depends_on:
  - redis
 web2:
  restart: on-failure
  build: ./web
  hostname: web2
  ports:
    - '82:5000'
 nginx:
  build: ./nginx
  ports:
    - '80:80'
  depends_on:
  - web1
  - web2
`)
	_, err = convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.Error(t, err)
	require.ErrorIs(t, CyclicalDependencyError, err)
}

// ====================================================================================================
//
//	Tests for  docker-compose files in awesome-compose (https://github.com/docker/awesome-compose)
//
// ====================================================================================================
// https://github.com/docker/awesome-compose/tree/master/minecraft
func TestMinecraftCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

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
    plan.add_service(name = "minecraft", config = ServiceConfig(image="itzg/minecraft-server", ports={"port0": PortSpec(number=25565, transport_protocol="TCP")}, files={"/data": Directory(persistent_key="minecraft--volume0")}, env_vars={"EULA": "TRUE"}, min_cpu=0, min_memory=0))
`

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// https://github.com/docker/awesome-compose/tree/master/angular
func TestAngularCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	relVolumePath := "./angular"
	err = os.Mkdir(path.Join(testPackageAbsDirPath, relVolumePath), testFilePerms)
	require.Nil(t, err)

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
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.upload_files(src = "./angular", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%v", build_context_dir="angular", target_stage="builder"), ports={"port0": PortSpec(number=4200, transport_protocol="TCP")}, files={"/project": "web--volume0", "/project/node_modules": Directory(persistent_key="web--volume1")}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/elasticsearch-logstash-kibana
func TestElasticSearchLogStashAndKibanaCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	logstashDirPath := "./logstash/pipeline"
	err = os.MkdirAll(path.Join(testPackageAbsDirPath, logstashDirPath), 0750)
	require.Nil(t, err)
	nginxConfigRelFilePath := "./logstash/pipeline/logstash-nginx.config"
	_, err = os.Create(path.Join(testPackageAbsDirPath, nginxConfigRelFilePath))
	require.Nil(t, err)
	nginxLogRelFilePath := "./logstash/nginx.log"
	_, err = os.Create(path.Join(testPackageAbsDirPath, nginxLogRelFilePath))
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  elasticsearch:
    image: elasticsearch:7.16.1
    container_name: es
    environment:
      discovery.type: single-node
      ES_JAVA_OPTS: "-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
      - "9300:9300"
    healthcheck:
      test: ["CMD-SHELL", "curl --silent --fail localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 3
    networks:
      - elastic
  logstash:
    image: logstash:7.16.1
    container_name: log
    environment:
      discovery.seed_hosts: logstash
      LS_JAVA_OPTS: "-Xms512m -Xmx512m"
    volumes:
      - ./logstash/pipeline/logstash-nginx.config:/usr/share/logstash/pipeline/logstash-nginx.config
      - ./logstash/nginx.log:/home/nginx.log
    ports:
      - "5000:5000/tcp"
      - "5000:5000/udp"
      - "5044:5044"
      - "9600:9600"
    depends_on:
      - elasticsearch
    networks:
      - elastic
    command: logstash -f /usr/share/logstash/pipeline/logstash-nginx.config
  kibana:
    image: kibana:7.16.1
    container_name: kib
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch
    networks:
      - elastic
networks:
  elastic:
    driver: bridge
`)

	// service names are equal to container names
	// files_be_moved added to service config to handle mounting files specifically
	expectedResult := `def run(plan):
    plan.add_service(name = "es", config = ServiceConfig(image="elasticsearch:7.16.1", ports={"port0": PortSpec(number=9200, transport_protocol="TCP"), "port1": PortSpec(number=9300, transport_protocol="TCP")}, env_vars={"ES_JAVA_OPTS": "-Xms512m -Xmx512m", "discovery.type": "single-node"}))
    plan.add_service(name = "kib", config = ServiceConfig(image="kibana:7.16.1", ports={"port0": PortSpec(number=5601, transport_protocol="TCP")}, env_vars={}))
    plan.upload_files(src = "./logstash/nginx.log", name = "log--volume1")
    plan.upload_files(src = "./logstash/pipeline/logstash-nginx.config", name = "log--volume0")
    plan.add_service(name = "log", config = ServiceConfig(image="logstash:7.16.1", ports={"port0": PortSpec(number=5000, transport_protocol="TCP"), "port1": PortSpec(number=5000, transport_protocol="UDP"), "port2": PortSpec(number=5044, transport_protocol="TCP"), "port3": PortSpec(number=9600, transport_protocol="TCP")}, files={"/tmp/log--volume0": "log--volume0", "/tmp/log--volume1": "log--volume1"}, cmd=["logstash", "-f", "/usr/share/logstash/pipeline/logstash-nginx.config"], env_vars={"LS_JAVA_OPTS": "-Xms512m -Xmx512m", "discovery.seed_hosts": "logstash"}, files_to_be_moved={"/tmp/log--volume0/logstash-nginx.config": "/usr/share/logstash/pipeline/logstash-nginx.config", "/tmp/log--volume1/nginx.log": "/home/nginx.log"}))
`
	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/fastapi/compose.yaml
func TestFastApiCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  api:
    build:
      context: .
      target: builder
    environment:
      PORT: 8000
    ports:
      - '8000:8000'
    restart: "no"
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "api", config = ServiceConfig(image=ImageBuildSpec(image_name="api%v", build_context_dir=".", target_stage="builder"), ports={"port0": PortSpec(number=8000, transport_protocol="TCP", application_protocol="http", url="http://api:8000")}, env_vars={"PORT": "8000"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/flask-redis/compose.yaml
func TestFlaskRedisCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)
	relVolumePath := "./code"
	err = os.Mkdir(path.Join(testPackageAbsDirPath, relVolumePath), testFilePerms)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  redis:
    image: redislabs/redismod
    ports:
      - '6379:6379'
  web:
    build:
      context: .
      target: builder
    # flask requires SIGINT to stop gracefully
    # (default stop signal from Compose is SIGTERM)
    stop_signal: SIGINT
    ports:
      - '8000:8000'
    volumes:
      - ./code:/code
    depends_on:
      - redis
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "redis", config = ServiceConfig(image="redislabs/redismod", ports={"port0": PortSpec(number=6379, transport_protocol="TCP")}, env_vars={}))
    plan.upload_files(src = "./code", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%v", build_context_dir=".", target_stage="builder"), ports={"port0": PortSpec(number=8000, transport_protocol="TCP", application_protocol="http", url="http://web:8000")}, files={"/code": "web--volume0"}, env_vars={}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/nextcloud-redis-mariadb/compose.yaml
func TestNextCloudRedisMariaDBCompose(t *testing.T) {
	testPackageAbsDirPath, err := os.MkdirTemp("", testPackageAbsDirPathPattern)
	defer os.RemoveAll(testPackageAbsDirPath)
	require.Nil(t, err)

	composeBytes := []byte(`
services:
  nc:
    image: nextcloud:apache
    restart: always
    ports:
      - 80:80
    volumes:
      - nc_data:/var/www/html
    networks:
      - redisnet
      - dbnet
    environment:
      - REDIS_HOST=redis
      - MYSQL_HOST=db
      - MYSQL_DATABASE=nextcloud
      - MYSQL_USER=nextcloud
      - MYSQL_PASSWORD=nextcloud
  redis:
    image: redis:alpine
    restart: always
    networks:
      - redisnet
    expose:
      - 6379
  db:
    image: mariadb:10.5
    command: --transaction-isolation=READ-COMMITTED --binlog-format=ROW
    restart: always
    volumes:
      - db_data:/var/lib/mysql
    networks:
      - dbnet
    environment:
      - MYSQL_DATABASE=nextcloud
      - MYSQL_USER=nextcloud
      - MYSQL_ROOT_PASSWORD=nextcloud
      - MYSQL_PASSWORD=nextcloud
    expose:
      - 3306
volumes:
  db_data:
  nc_data:
networks:
  dbnet:
  redisnet:
`)
	expectedResult := `def run(plan):
    plan.add_service(name = "db", config = ServiceConfig(image="mariadb:10.5", files={"/var/lib/mysql": Directory(persistent_key="db--volume0")}, cmd=["--transaction-isolation=READ-COMMITTED", "--binlog-format=ROW"], env_vars={"MYSQL_DATABASE": "nextcloud", "MYSQL_PASSWORD": "nextcloud", "MYSQL_ROOT_PASSWORD": "nextcloud", "MYSQL_USER": "nextcloud"}))
    plan.add_service(name = "nc", config = ServiceConfig(image="nextcloud:apache", ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://nc:80")}, files={"/var/www/html": Directory(persistent_key="nc--volume0")}, env_vars={"MYSQL_DATABASE": "nextcloud", "MYSQL_HOST": "db", "MYSQL_PASSWORD": "nextcloud", "MYSQL_USER": "nextcloud", "REDIS_HOST": "redis"}))
    plan.add_service(name = "redis", config = ServiceConfig(image="redis:alpine", env_vars={}))
`

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{}, testPackageAbsDirPath)
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/nginx-aspnet-mysql/compose.yaml

// From https://github.com/docker/awesome-compose/blob/master/nginx-flask-mongo/compose.yaml

// From https://github.com/docker/awesome-compose/blob/master/nginx-flask-mysql/compose.yaml/

// From https://github.com/docker/awesome-compose/blob/master/nginx-golang-mysql/compose.yaml

// ====================================================================================================
//
//	Tests from other docker composes in the wild
//
// ====================================================================================================

package docker_compose_transpiler

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO: Create a test framework like starlark test framework

func TestMinimalCompose(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithImageBuildSpec(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithImageBuildSpecAndTarget(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithVolume(t *testing.T) {
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
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/tmp/web--volume0": "web--volume0"}, env_vars={}, files_to_be_moved={"/tmp/web--volume0/data": "/data"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolume(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolumeAtProvidedPath(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMinimalComposeWithPersistentVolumeAtProvidedPathAreUnique(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// Tests all supported compose functionalities for a single service
func TestFullCompose(t *testing.T) {
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
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="app", target_stage="builder"), ports={"port0": PortSpec(number=80, transport_protocol="TCP", application_protocol="http", url="http://web:80")}, files={"/node_modules": Directory(persistent_key="web--volume1"), "/tmp/web--volume0": "web--volume0"}, entrypoint=["/bin/echo", "-c", "echo \"Hello\""], cmd=["echo", "Hello,", "World!"], env_vars={"NODE_ENV": "development"}, files_to_be_moved={"/tmp/web--volume0/data": "/data"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestMultiServiceCompose(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
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

func TestWasmedgeCompose(t *testing.T) {
	composeBytes := []byte(`
services:
  redpanda:
    image: docker.redpanda.com/vectorized/redpanda:v22.2.2
    command:
      - redpanda start
      - --smp 1
      - --overprovisioned
      - --node-id 0
      - --kafka-addr PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:9092
      - --advertise-kafka-addr PLAINTEXT://redpanda:29092,OUTSIDE://redpanda:9092
      - --pandaproxy-addr 0.0.0.0:8082
      - --advertise-pandaproxy-addr localhost:8082
    ports:
      - 8081:8081
      - 8082:8082
      - 9092:9092
      - 9644:9644
      - 29092:29092
    volumes:
      - ./kafka:/app
  etl:
    image: etl-kafka
    platform: wasi/wasm
    build:
      context: etl
    environment:
      DATABASE_URL: mysql://root:whalehello@db:3306/mysql
      KAFKA_URL: kafka://redpanda:9092/order
      RUST_BACKTRACE: full
      RUST_LOG: info
    restart: unless-stopped
    runtime: io.containerd.wasmedge.v1
  db:
    image: mariadb:10.9
    environment:
      MYSQL_ROOT_PASSWORD: whalehello`)

	_, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
}

func TestServiceName(t *testing.T) {
	composeBytes := []byte(`
version: "3.9"
services:
  blockscout:
    depends_on:
      - postgres
      - sequencer
    image: blockscout-testnode
    restart: always
    container_name: 'blockscout'
    links:
      - postgres:database
    command:
      - /bin/sh
      - -c
      - |
        bin/blockscout eval "Elixir.Explorer.ReleaseTasks.create_and_migrate()"
        node init/install.js postgres 5432
        bin/blockscout start
    extra_hosts:
      - 'host.docker.internal:host-gateway'
    environment:
        ETHEREUM_JSONRPC_VARIANT: 'geth'
        ETHEREUM_JSONRPC_HTTP_URL: http://sequencer:8547/
        INDEXER_DISABLE_PENDING_TRANSACTIONS_FETCHER: "true"
        DATABASE_URL: postgresql://postgres:@postgres:5432/blockscout
        ECTO_USE_SSL: "false"
    ports:
      - "127.0.0.1:4000:4000"

  postgres:
    image: postgres:13.6
    restart: always
    container_name: 'postgres'
    environment:
        POSTGRES_PASSWORD: ''
        POSTGRES_USER: 'postgres'
        POSTGRES_HOST_AUTH_METHOD: 'trust'
    volumes:
      - "postgres-data:/var/lib/postgresql/data"
    ports:
      - "127.0.0.1:7432:5432"

  redis:
    image: redis:6.2.6
    ports:
      - "127.0.0.1:6379:6379"

  geth:
    image: ethereum/client-go:v1.10.23
    ports:
      - "127.0.0.1:8545:8545"
      - "127.0.0.1:8551:8551"
      - "127.0.0.1:8546:8546"
      - "127.0.0.1:30303:30303"
    volumes:
      - "l1data:/datadir"
      - "l1keystore:/keystore"
      - "config:/config"
    command:
      - --keystore=/keystore
      - --http
      - --datadir=/datadir
      - --http.addr=0.0.0.0
      - --authrpc.vhosts=*
      - --authrpc.port=8551
      - --authrpc.addr=0.0.0.0
      - --http.vhosts=*
      - --http.api=engine,personal,eth,net,web3
      - --http.corsdomain=*
      - --ws
      - --ws.addr=0.0.0.0
      - --ws.api=personal,eth,net,web3,debug,txpool
      - --allow-insecure-unlock
      - --unlock=0x3f1Eae7D46d88F08fc2F8ed27FCb2AB183EB2d0E
      - --password=/datadir/passphrase
      - --authrpc.jwtsecret=/config/jwt.hex
      - --nodiscover
      - --syncmode=full
      - --dev
      - --dev.period=1
      - --mine
      - --miner.etherbase=0x3f1Eae7D46d88F08fc2F8ed27FCb2AB183EB2d0E
      - --gcmode=archive

  # Creates a genesis state for the beacon chain using a YAML configuration file and
  # a deterministic set of validators
  # TODO: Make num validators customizable
  create_beacon_chain_genesis:
    image: "gcr.io/prysmaticlabs/prysm/cmd/prysmctl:latest"
    command:
      - testnet
      - generate-genesis
      - --num-validators=64
      - --output-ssz=/consensus/genesis.ssz
      - --chain-config-file=/config/prysm.yaml
    volumes:
      - "consensus:/consensus"
      - "config:/config"

  # Runs a Prysm beacon chain from a specified genesis state created in the previous step
  # and connects to go-ethereum in the same network as the execution client.
  # The account used in go-ethereum is set as the suggested fee recipient for transactions
  # proposed via the validators attached to the beacon node.
  prysm_beacon_chain:
    image: "gcr.io/prysmaticlabs/prysm/beacon-chain:stable"
    command:
      - --datadir=/consensus/beacondata
      - --rpc-port=5000
      - --min-sync-peers=0
      - --interop-genesis-state=/consensus/genesis.ssz
      - --interop-eth1data-votes
      - --bootstrap-node=
      - --chain-config-file=/config/prysm.yaml
      - --rpc-host=0.0.0.0
      - --grpc-gateway-host=0.0.0.0
      - --chain-id=32382
      - --execution-endpoint=http://geth:8551
      - --accept-terms-of-use
      - --jwt-secret=/config/jwt.hex
    depends_on:
      geth:
        condition: service_started
      create_beacon_chain_genesis:
        condition: service_completed_successfully
    ports:
      - "127.0.0.1:5000:5000"
      - "127.0.0.1:3500:3500"
    volumes:
      - "consensus:/consensus"
      - "config:/config"

  # We run a validator client with 64, deterministically-generated keys that match
  # The validator keys present in the beacon chain genesis state generated a few steps above.
  prysm_validator:
    image: "gcr.io/prysmaticlabs/prysm/validator:stable"
    command:
      - --beacon-rpc-provider=prysm_beacon_chain:5000
      - --datadir=/consensus/validatordata
      - --accept-terms-of-use
      - --interop-num-validators=64
      - --interop-start-index=0
      - --chain-config-file=/config/prysm.yaml
    depends_on:
      prysm_beacon_chain:
        condition: service_started
    volumes:
      - "consensus:/consensus"
      - "config:/config"

  sequencer:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8547:8547"
      - "127.0.0.1:8548:8548"
      - "127.0.0.1:9642:9642"
    volumes:
      - "seqdata:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
      - "tokenbridge-data:/tokenbridge-data"
    command: --conf.file /config/sequencer_config.json --node.feed.output.enable --node.feed.output.port 9642  --http.api net,web3,eth,txpool,debug --node.seq-coordinator.my-url  ws://sequencer:8548 --graphql.enable --graphql.vhosts * --graphql.corsdomain *
    depends_on:
      - geth

  sequencer_b:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8647:8547"
      - "127.0.0.1:8648:8548"
    volumes:
      - "seqdata_b:/home/user/.arbitrum/local/nitro"
      - "config:/config"
    command: --conf.file /config/sequencer_config.json --node.seq-coordinator.my-url ws://sequencer_b:8548
    depends_on:
      - geth
      - redis

  sequencer_c:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8747:8547"
      - "127.0.0.1:8748:8548"
    volumes:
      - "seqdata_c:/home/user/.arbitrum/local/nitro"
      - "config:/config"
    command: --conf.file /config/sequencer_config.json --node.seq-coordinator.my-url ws://sequencer_c:8548
    depends_on:
      - geth
      - redis

  sequencer_d:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8847:8547"
      - "127.0.0.1:8848:8548"
    volumes:
      - "seqdata_d:/home/user/.arbitrum/local/nitro"
      - "config:/config"
    command: --conf.file /config/sequencer_config.json --node.seq-coordinator.my-url ws://sequencer_d:8548
    depends_on:
      - geth
      - redis

  staker-unsafe:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8047:8547"
      - "127.0.0.1:8048:8548"
    volumes:
      - "unsafestaker-data:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/unsafe_staker_config.json
    depends_on:
      - sequencer
      - redis
      - validation_node

  poster:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8147:8547"
      - "127.0.0.1:8148:8548"
    volumes:
      - "poster-data:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/poster_config.json
    depends_on:
      - geth
      - redis

  poster_b:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:9147:8547"
      - "127.0.0.1:9148:8548"
    volumes:
      - "poster-data-b:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/poster_config.json
    depends_on:
      - geth
      - redis

  poster_c:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:9247:8547"
      - "127.0.0.1:9248:8548"
    volumes:
      - "poster-data-c:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/poster_config.json
    depends_on:
      - geth
      - redis

  validator:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8247:8547"
      - "127.0.0.1:8248:8548"
    volumes:
      - "validator-data:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/validator_config.json --http.port 8547 --http.api net,web3,arb,debug --ws.port 8548
    depends_on:
      - sequencer
      - validation_node

  l3node:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:3347:3347"
      - "127.0.0.1:3348:3348"
    volumes:
      - "validator-data:/home/user/.arbitrum/local/nitro"
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
    command: --conf.file /config/l3node_config.json --http.port 3347 --http.api net,web3,arb,debug,eth --ws.port 3348
    depends_on:
      - sequencer
      - validation_node

  validation_node:
    pid: host # allow debugging
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:8949:8549"
    volumes:
      - "config:/config"
    command: --conf.file /config/validation_node_config.json
    entrypoint: /usr/local/bin/nitro-val

  scripts:
    build: scripts/
    volumes:
      - "l1keystore:/home/user/l1keystore"
      - "config:/config"
      - "tokenbridge-data:/tokenbridge-data"

  relay:
    pid: host
    image: nitro-node-dev-testnode
    ports:
      - "127.0.0.1:9652:9652"
    entrypoint: bin/relay
    command: --node.feed.output.port 9652 --node.feed.input.url ws://sequencer:9652

  tokenbridge:
    depends_on:
      - geth
      - sequencer
    pid: host
    build: tokenbridge/
    environment:
      - ARB_URL=http://sequencer:8547
      - ETH_URL=http://geth:8545
    volumes:
      - "tokenbridge-data:/workspace"
      - /var/run/docker.sock:/var/run/docker.sock

volumes:
  l1data:
  consensus:
  l1keystore:
  seqdata:
  seqdata_b:
  seqdata_c:
  seqdata_d:
  unsafestaker-data:
  validator-data:
  poster-data:
  poster-data-b:
  poster-data-c:
  config:
  postgres-data:
  tokenbridge-data:
`)

	_, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
}

func TestMultiServiceComposeWithDependsOn(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

//
//func TestServiceLevelEnvFile(t *testing.T) {
//	envFileData := "DB=kurtosis\nUSER=kurtosian"
//	envFilePath := "./web.env"
//
//	err := os.Mkdir("/kurtosis-data", 0644)
//	require.Nil(t, err)
//	defer os.RemoveAll("/kurtosis-data")
//	err = os.Mkdir("/repositories", 0644)
//	require.Nil(t, err)
//	err = os.Mkdir("/NOTIONAL_USER", 0644)
//	require.Nil(t, err)
//	err = os.Mkdir("/USER_UPLOADED_COMPOSE_PACKAGE", 0644)
//	require.Nil(t, err)
//
//	tmpEnvFilePath := path.Join(serviceLevelEnvFileDirPath, envFilePath)
//	_, err = os.Create(tmpEnvFilePath)
//	require.NoError(t, err)
//	err = os.WriteFile(tmpEnvFilePath, []byte(envFileData), 0666)
//	require.NoError(t, err)
//
//	composeBytes := []byte(`
//services:
//  web:
//    image: app/server
//    env_file:
//     - ./ ` + envFilePath + `
//`)
//
//	expectedResult := `def run(plan):
//    plan.add_service(name = "web", config = ServiceConfig(image="app/server", env_vars={"DB": "kurtosis", "USER": "kurtosian"}))
//`
//	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
//	require.NoError(t, err)
//	require.Equal(t, expectedResult, result)
//}

// Test depends on with circular dependency returns error
func TestMultiServiceComposeWithCycleInDependsOn(t *testing.T) {
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
	_, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
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
    plan.add_service(name = "minecraft", config = ServiceConfig(image="itzg/minecraft-server", ports={"port0": PortSpec(number=25565, transport_protocol="TCP")}, files={"/data": Directory(persistent_key="minecraftdata")}, env_vars={"EULA": "TRUE"}, min_cpu=0, min_memory=0))
`

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// https://github.com/docker/awesome-compose/tree/master/angular
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
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.upload_files(src = "./angular", name = "web--volume0")
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%s", build_context_dir="angular", target_stage="builder"), ports={"port0": PortSpec(number=4200, transport_protocol="TCP")}, files={"/project/node_modules": Directory(persistent_key="web--volume1"), "/tmp/web--volume0": "web--volume0"}, env_vars={}, files_to_be_moved={"/tmp/web--volume0/angular": "/project"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/elasticsearch-logstash-kibana
func TestElasticSearchLogStashAndKibanaCompose(t *testing.T) {
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
	expectedResult := `def run(plan):
    plan.add_service(name = "elasticsearch", config = ServiceConfig(image="elasticsearch:7.16.1", ports={"port0": PortSpec(number=9200, transport_protocol="TCP"), "port1": PortSpec(number=9300, transport_protocol="TCP")}, env_vars={"ES_JAVA_OPTS": "-Xms512m -Xmx512m", "discovery.type": "single-node"}))
    plan.add_service(name = "kibana", config = ServiceConfig(image="kibana:7.16.1", ports={"port0": PortSpec(number=5601, transport_protocol="TCP")}, env_vars={}))
    plan.upload_files(src = "./logstash/nginx.log", name = "logstash--volume1")
    plan.upload_files(src = "./logstash/pipeline/logstash-nginx.config", name = "logstash--volume0")
    plan.add_service(name = "logstash", config = ServiceConfig(image="logstash:7.16.1", ports={"port0": PortSpec(number=5000, transport_protocol="TCP"), "port1": PortSpec(number=5000, transport_protocol="UDP"), "port2": PortSpec(number=5044, transport_protocol="TCP"), "port3": PortSpec(number=9600, transport_protocol="TCP")}, files={"/tmp/logstash--volume0": "logstash--volume0", "/tmp/logstash--volume1": "logstash--volume1"}, cmd=["logstash", "-f", "/usr/share/logstash/pipeline/logstash-nginx.config"], env_vars={"LS_JAVA_OPTS": "-Xms512m -Xmx512m", "discovery.seed_hosts": "logstash"}, files_to_be_moved={"/tmp/logstash--volume0/logstash-nginx.config": "/usr/share/logstash/pipeline/logstash-nginx.config", "/tmp/logstash--volume1/nginx.log": "/home/nginx.log"}))
`

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/fastapi/compose.yaml
func TestFastApiCompose(t *testing.T) {
	composeBytes := []byte(`
services:
  api:
    build:
      context: .
      target: builder
    container_name: fastapi-application
    environment:
      PORT: 8000
    ports:
      - '8000:8000'
    restart: "no"
`)
	expectedResult := fmt.Sprintf(`def run(plan):
    plan.add_service(name = "api", config = ServiceConfig(image=ImageBuildSpec(image_name="api%v", build_context_dir=".", target_stage="builder"), ports={"port0": PortSpec(number=8000, transport_protocol="TCP", application_protocol="http", url="http://api:8000")}, env_vars={"PORT": "8000"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/flask-redis/compose.yaml
func TestFlaskRedisCompose(t *testing.T) {
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
    plan.add_service(name = "web", config = ServiceConfig(image=ImageBuildSpec(image_name="web%v", build_context_dir=".", target_stage="builder"), ports={"port0": PortSpec(number=8000, transport_protocol="TCP", application_protocol="http", url="http://web:8000")}, files={"/tmp/web--volume0": "web--volume0"}, env_vars={}, files_to_be_moved={"/tmp/web--volume0/code": "/code"}))
`, builtImageSuffix)

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

// From https://github.com/docker/awesome-compose/blob/master/nextcloud-redis-mariadb/compose.yaml
func TestNextCloudRedisMariaDBCompose(t *testing.T) {
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

	result, err := convertComposeToStarlarkScript(composeBytes, map[string]string{})
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

// TODO: Test this docker compose when named volumes are supported https://github.com/OffchainLabs/nitro-testnode/blob/release/docker-compose.yaml

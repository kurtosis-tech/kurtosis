package ava_commons

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/gmarchetti/kurtosis/commons/testnet"
	"strconv"
	"strings"
)

const (
	httpPort = 9650
	stakingPort = 9651
)

// ================= Gecko Service ==================================

type GeckoService struct {
	ipAddr string
}

func NewGeckoService(ipAddr string) *GeckoService {
	return &GeckoService{
		ipAddr:      ipAddr,
	}
}

func (g GeckoService) GetStakingSocket() testnet.ServiceSocket {
	stakingPort, err := nat.NewPort("tcp", strconv.Itoa(stakingPort))
	if err != nil {
		// Realllllly don't think we should deal with propagating this one.... it means the user mistyped an integer
		panic(err)
	}
	return *testnet.NewServiceSocket(g.ipAddr, stakingPort)
}

// TODO implement a GetJsonRpcSocket function, which we'll need for testing

// ================ Gecko Service Factory =============================
type geckoLogLevel string
const (
	LOG_LEVEL_VERBOSE geckoLogLevel = "verbo"
	LOG_LEVEL_DEBUG geckoLogLevel = "debug"
	LOG_LEVEL_INFO geckoLogLevel = "info"
)

type GeckoServiceFactoryConfig struct {
	dockerImage string
	snowSampleSize int
	snowQuorumSize int
	stakingTlsEnabled bool
	logLevel geckoLogLevel
}

func NewGeckoServiceFactoryConfig(dockerImage string,
			snowSampleSize int,
			snowQuorumSize int,
			stakingTlsEnabled bool,
			logLevel geckoLogLevel) *GeckoServiceFactoryConfig {
	return &GeckoServiceFactoryConfig{
		dockerImage:       dockerImage,
		snowSampleSize:    snowSampleSize,
		snowQuorumSize:    snowQuorumSize,
		stakingTlsEnabled: stakingTlsEnabled,
		logLevel:          logLevel,
	}
}

func (g GeckoServiceFactoryConfig) GetDockerImage() string {
	return g.dockerImage
}

func (g GeckoServiceFactoryConfig) GetUsedPorts() map[int]bool {
	return map[int]bool{
		httpPort: true,
		stakingPort: true,
	}
}

func (g GeckoServiceFactoryConfig) GetStartCommand(ipAddrOffset int, dependencies []testnet.Service) []string {
	commandList := []string{
		"/gecko/build/ava",
		// TODO this entire flag will go away soon!!
		fmt.Sprintf("--public-ip=172.17.0.%d", 2 + ipAddrOffset),
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", httpPort),
		fmt.Sprintf("--staking-port=%d", stakingPort),
		fmt.Sprintf("--log-level=%s", g.logLevel),
		fmt.Sprintf("--snow-sample-size=%d", g.snowSampleSize),
		fmt.Sprintf("--snow-quorum-size=%d", g.snowQuorumSize),
		fmt.Sprintf("--staking-tls-enabled=%v", g.stakingTlsEnabled),
	}

	// If bootstrap nodes are down then Gecko will wait until they are, so we don't actually need to busy-loop making
	// requests to the nodes
	if dependencies != nil && len(dependencies) > 0 {
		// TODO realllllllly wish Go had generics, so we didn't have to do this!
		avaDependencies := make([]AvaService, 0, len(dependencies))
		for _, service := range dependencies {
			println(fmt.Sprintf("Service: %v", service))
			avaDependencies = append(avaDependencies, service.(AvaService))
		}

		socketStrs := make([]string, 0, len(avaDependencies))
		for _, service := range avaDependencies {
			socket := service.GetStakingSocket()
			socketStrs = append(socketStrs, fmt.Sprintf("%s:%d", socket.GetIpAddr(), socket.GetPort().Int()))
		}
		joinedSockets := strings.Join(socketStrs, ",")
		commandList = append(commandList, "--bootstrap-ips=" + joinedSockets)
	}

	return commandList
}

func (g GeckoServiceFactoryConfig) GetServiceFromIp(ipAddr string) testnet.Service {
	return GeckoService{ipAddr: ipAddr}
}


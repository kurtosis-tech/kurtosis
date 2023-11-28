package port_forward_manager

import (
	"context"
	chclient "github.com/jpillora/chisel/client"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	remoteSeparatorString              = ":"
	chiselClientConfigKeepAlive        = 25 * time.Second
	chiselClientConfigMaxRetry         = -1 // unlimited retries
	chiselClientConfigMaxRetryInterval = 10 * time.Second
)

type PortForwardTunnel struct {
	localPortNumber        uint16
	serviceInterfaceDetail *ServiceInterfaceDetail

	context    context.Context
	cancelFunc context.CancelFunc

	chiselClient *chclient.Client
}

func NewPortForwardTunnel(localPortNumber uint16, sid *ServiceInterfaceDetail) (*PortForwardTunnel, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	remoteTunnelString := getRemoteTunnelString(localPortNumber, sid.serviceIpAddress, sid.servicePortSpec.GetNumber(), sid.servicePortSpec.GetTransportProtocol())
	chiselClient, err := newChiselClient(sid.chiselServerUri, remoteTunnelString)
	if err != nil {
		defer cancelFunc()
		return nil, stacktrace.Propagate(err, "Failed to create chisel tunnel to chisel server '%v' with remote spec '%v'", sid.chiselServerUri, remoteTunnelString)
	}

	return &PortForwardTunnel{
		localPortNumber,
		sid,

		ctx,
		cancelFunc,

		chiselClient,
	}, nil
}

// TODO(omar): lifecycle, locking, more error handling, etc
func (tunnel *PortForwardTunnel) RunAsync() error {
	if err := tunnel.chiselClient.Start(tunnel.context); err != nil {
		return stacktrace.Propagate(err, "Failed to start Chisel client")
	}
	return nil
}

func (tunnel *PortForwardTunnel) Close() {
	tunnel.cancelFunc()

	if tunnel.chiselClient != nil {
		if err := tunnel.chiselClient.Close(); err != nil {
			logrus.Warnf("Error encountered closing port tunneling client: \n%v", err.Error())
		}
	}
}

func getRemoteTunnelString(localPortNumber uint16, remoteServiceIp string, remotePortNumber uint16, transportProtocol services.TransportProtocol) string {
	remoteSpec := []string{
		strconv.Itoa(int(localPortNumber)),
		remoteServiceIp,
		strconv.Itoa(int(remotePortNumber)),
	}

	remoteString := strings.Join(remoteSpec, remoteSeparatorString)

	if transportProtocol == services.TransportProtocol_UDP {
		remoteString = remoteString + "/udp"
	}

	return remoteString
}

func newChiselClient(chiselServerUri string, remoteTunnelString string) (*chclient.Client, error) {
	chiselClientConfig := &chclient.Config{
		Fingerprint:      "",
		Auth:             "",
		KeepAlive:        chiselClientConfigKeepAlive,
		MaxRetryCount:    chiselClientConfigMaxRetry,
		MaxRetryInterval: chiselClientConfigMaxRetryInterval,
		Server:           chiselServerUri,
		Proxy:            "",
		Remotes: []string{
			remoteTunnelString,
		},
		Headers: nil,

		DialContext: nil,
		Verbose:     true,

		TLS: chclient.TLSConfig{
			SkipVerify: true,
			CA:         "",
			Cert:       "",
			Key:        "",
			ServerName: "",
		},
	}

	return chclient.NewClient(chiselClientConfig)
}

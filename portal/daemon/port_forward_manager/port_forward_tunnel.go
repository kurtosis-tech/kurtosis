package port_forward_manager

import (
	"context"
	chclient "github.com/jpillora/chisel/client"
	"github.com/kurtosis-tech/stacktrace"
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
	localPortNumber   uint16
	remoteServiceIp   string
	remoteServicePort uint16
	chiselServerUri   string

	context    context.Context
	cancelFunc context.CancelFunc
}

func NewPortForwardTunnel(localPortNumber uint16, sid *ServiceInterfaceDetail) *PortForwardTunnel {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &PortForwardTunnel{
		localPortNumber,
		sid.serviceIpAddress,
		sid.servicePortSpec.GetNumber(),
		sid.chiselServerUri,

		ctx,
		cancelFunc,
	}
}

// TODO(omar): lifecycle, locking, more error handling, etc
func (session *PortForwardTunnel) RunAsync() error {
	remoteTunnelString := session.getRemoteTunnelString()
	chiselClient, err := session.getChiselClient(remoteTunnelString)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create chisel tunnel to chisel server '%v' with remote spec '%v'", session.chiselServerUri, remoteTunnelString)
	}

	if err := chiselClient.Start(session.context); err != nil {
		return stacktrace.Propagate(err, "Unable to start Chisel client for remote: '%s'", remoteTunnelString)
	}
	return nil
}

func (tunnel *PortForwardTunnel) getRemoteTunnelString() string {
	remoteSpec := []string{
		strconv.Itoa(int(tunnel.localPortNumber)),
		tunnel.remoteServiceIp,
		strconv.Itoa(int(tunnel.remoteServicePort)),
	}
	return strings.Join(remoteSpec, remoteSeparatorString)
}

func (tunnel *PortForwardTunnel) getChiselClient(remoteTunnelString string) (*chclient.Client, error) {
	chiselClientConfig := &chclient.Config{
		Fingerprint:      "",
		Auth:             "",
		KeepAlive:        chiselClientConfigKeepAlive,
		MaxRetryCount:    chiselClientConfigMaxRetry,
		MaxRetryInterval: chiselClientConfigMaxRetryInterval,
		Server:           tunnel.chiselServerUri,
		Proxy:            "",
		Remotes: []string{
			remoteTunnelString,
		},
		Headers: nil,

		DialContext: nil,
		Verbose:     true,
	}

	return chclient.NewClient(chiselClientConfig)
}

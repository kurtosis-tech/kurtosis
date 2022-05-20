package connection

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	k8s_rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"time"
)

const (
	localHostIpStr             = "127.0.0.1"
	portForwardTimeoutDuration = 5 * time.Second
	grpcPortId                 = "grpc"
)

//GatewayConnectionToKurtosis represents a connection on localhost that can be used by the gateway to communicate with Kurtosis in the cluster
type GatewayConnectionToKurtosis interface {
	// GetLocalPorts returns a map keyed with an identifier string describing local ports being forwarded
	GetLocalPorts() map[string]*port_spec.PortSpec
	GetGrpcClientConn() (*grpc.ClientConn, error)
	Stop()
}

// A gateway connection using `kubectl proxy`
type gatewayConnectionToKurtosisImpl struct {
	localAddresses []string
	localPorts     map[string]*port_spec.PortSpec
	// kubectl command to exec to to start the tunnel
	portforwarder *portforward.PortForwarder

	portforwarderStdOut bytes.Buffer
	portforwarderStdErr bytes.Buffer

	portforwarderStopChannel  chan struct{}
	portforwarderReadyChannel chan struct{}

	// RemotePort -> port-spec ID
	remotePortNumberToPortSpecIdMap map[uint16]string

	urlString string
}

// newLocalPortToPodPortConnection binds a random local port to the remote port keyed with an identifier string
// remotePortSpecs is a map keyed with an identifier string of port specs on the remote pod to forward requests to
func newLocalPortToPodPortConnection(kubernetesRestConfig *k8s_rest.Config, podProxyEndpointUrl *url.URL, remotePortSpecs map[string]*port_spec.PortSpec) (*gatewayConnectionToKurtosisImpl, error) {
	var portforwardStdOut bytes.Buffer
	var portforwardStdErr bytes.Buffer
	portforwardStopChannel := make(chan struct{}, 1)
	portforwardReadyChannel := make(chan struct{}, 1)
	portForwardAddresses := []string{localHostIpStr}
	remotePortNumberToPortSpecIdMapping := map[uint16]string{}

	// Array of strings describing local-port:remote-port bindings
	portStrings := []string{}
	for portspecId, portSpec := range remotePortSpecs {
		// Kubernetes port-forwarding currently only supports TCP
		// https://github.com/kubernetes/kubernetes/issues/47862
		if portSpec.GetProtocol() != port_spec.PortProtocol_TCP {
			// Warn the user the this port won't be forwarded
			logrus.Warnf("The port with id '%v' won't be able to be forwarded from Kubernetes, it uses protocol '%v', but Kubernetes port forwarding only support the '%v' protocol", portspecId, portSpec.GetProtocol(), port_spec.PortProtocol_TCP)
			continue
		}
		// For our purposes, local-port is set to 0, meaning the host will assign us a random local port
		portString := fmt.Sprintf("0:%v", portSpec.GetNumber())
		portStrings = append(portStrings, portString)
		// Keep track of the portspec ID for the remote ports we connect to
		remotePortNumberToPortSpecIdMapping[portSpec.GetNumber()] = portspecId
	}

	// Connection to pod portforwarder endpoint
	transport, upgrader, err := spdy.RoundTripperFor(kubernetesRestConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a SPDY round-tripper for the Kubernetes REST config")
	}
	dialerMethod := "POST"
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, dialerMethod, podProxyEndpointUrl)

	// Create our k8s port forwarder
	portForwarder, err := portforward.NewOnAddresses(dialer, portForwardAddresses, portStrings, portforwardStopChannel, portforwardReadyChannel, &portforwardStdOut, &portforwardStdErr)
	if err != nil {
		return nil, err
	}
	// Start forwarding ports asynchronously
	go func() {
		if err := portForwarder.ForwardPorts(); err != nil {
			logrus.Warnf("Expected to be able to start forwarding local ports to remote ports, instead our portforwarder has returned a non-nil err:\n%v", err)
		}
	}()
	// Wait for the portforwarder to be ready with timeout
	select {
	case _ = <-portforwardReadyChannel:
	case <-time.After(portForwardTimeoutDuration):
		return nil, stacktrace.NewError("Expected Kubernetes portforwarder to open local ports to the pod exposed by the portforward api at URL '%v', instead the Kubernetes portforwarder timed out binding local ports", podProxyEndpointUrl)
	}
	// Get local forwarded ports
	forwardedPorts, err := portForwarder.GetPorts()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get forwarded ports from our running portforwarder, instead a non-nil err was returned")
	}
	// Get port specs and ids for local ports
	localPortSpecs := map[string]*port_spec.PortSpec{}
	for _, forwardedPort := range forwardedPorts {
		remotePort := forwardedPort.Remote
		localPort := forwardedPort.Local
		portSpecId, isFound := remotePortNumberToPortSpecIdMapping[remotePort]
		if !isFound {
			return nil, stacktrace.NewError("Expected to be able to find port_spec id of remote port '%v', instead found nothing", remotePort)
		}
		// Port forwarding in kubernetes only supports TCP
		localPortSpec, err := port_spec.NewPortSpec(localPort, port_spec.PortProtocol_TCP)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to create port-spec describing local port '%v', instead a non-nil err was returned", localPort)
		}

		logrus.Infof("Forwarding requests to local port number '%v' to remote port with id '%v'", localPortSpec.GetNumber(), portSpecId)

		localPortSpecs[portSpecId] = localPortSpec
	}

	connection := &gatewayConnectionToKurtosisImpl{
		localAddresses:            portForwardAddresses,
		localPorts:                localPortSpecs,
		portforwarder:             portForwarder,
		portforwarderStdOut:       portforwardStdOut,
		portforwarderStdErr:       portforwardStdErr,
		portforwarderStopChannel:  portforwardStopChannel,
		portforwarderReadyChannel: portforwardReadyChannel,
		urlString:                 podProxyEndpointUrl.String(),
	}

	return connection, nil
}

func (connection *gatewayConnectionToKurtosisImpl) Stop() {
	connection.portforwarder.Close()
	// stopping the channel is necessary to close the connection
	close(connection.portforwarderStopChannel)
}

func (connection *gatewayConnectionToKurtosisImpl) GetLocalPorts() map[string]*port_spec.PortSpec {
	return connection.localPorts
}

// GetGrpcClientConn returns a client conn dialed in to the local port
// It is the caller's responsibility to call resultClientConn.close()
func (connection *gatewayConnectionToKurtosisImpl) GetGrpcClientConn() (resultClientConn *grpc.ClientConn, resultErr error) {
	localPorts := connection.GetLocalPorts()
	localGrpcPort, isFound := localPorts[grpcPortId]
	if !isFound {
		return nil, stacktrace.NewError("Expected to find port_spec with id '%v', instead found nil", localGrpcPort)
	}
	localGrpcPortNum := localPorts[grpcPortId].GetNumber()
	localGrpcServerAddress := fmt.Sprintf("%v:%v", localHostIpStr, localGrpcPortNum)
	grpcConnection, err := grpc.Dial(localGrpcServerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a GRPC client connection on address '%v', but a non-nil error was returned", localGrpcServerAddress)
	}

	return grpcConnection, nil
}

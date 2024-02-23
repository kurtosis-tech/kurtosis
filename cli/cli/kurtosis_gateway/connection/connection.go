package connection

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	k8s_rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"time"
)

const (
	localHostIpStr                = "127.0.0.1"
	portForwardTimeoutDuration    = 5 * time.Second
	grpcPortId                    = "grpc"
	emptyApplicationProtocol      = ""
	portForwardTimeBetweenRetries = 5 * time.Second
)

// GatewayConnectionToKurtosis represents a connection on localhost that can be used by the gateway to communicate with Kurtosis in the cluster
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

	portforwarder            *portforward.PortForwarder
	portforwarderStopChannel chan struct{}

	stopChannel chan struct{}

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
	stopChannel := make(chan struct{}, 1)
	portForwardAddresses := []string{localHostIpStr}
	remotePortNumberToPortSpecIdMapping := map[uint16]string{}

	// Array of strings describing local-port:remote-port bindings
	portStrings := []string{}
	for portspecId, portSpec := range remotePortSpecs {
		// Kubernetes port-forwarding currently only supports TCP
		// https://github.com/kubernetes/kubernetes/issues/47862
		if portSpec.GetTransportProtocol() != port_spec.TransportProtocol_TCP {
			// Warn the user the this port won't be forwarded
			logrus.Warnf("The port with id '%v' won't be able to be forwarded from Kubernetes, it uses protocol '%v', but Kubernetes port forwarding only support the '%v' protocol", portspecId, portSpec.GetTransportProtocol(), port_spec.TransportProtocol_TCP)
			continue
		}
		// For our purposes, local-port is set to 0, meaning the host will assign us a random local port
		portString := fmt.Sprintf("0:%v", portSpec.GetNumber())
		portStrings = append(portStrings, portString)
		// Keep track of the portspec ID for the remote ports we connect to
		remotePortNumberToPortSpecIdMapping[portSpec.GetNumber()] = portspecId
	}

	connection := &gatewayConnectionToKurtosisImpl{
		localAddresses:                  portForwardAddresses,
		localPorts:                      nil,
		portforwarder:                   nil,
		portforwarderStopChannel:        portforwardStopChannel,
		stopChannel:                     stopChannel,
		remotePortNumberToPortSpecIdMap: remotePortNumberToPortSpecIdMapping,
		urlString:                       podProxyEndpointUrl.String(),
	}

	// Connection to pod portforwarder endpoint
	transport, upgrader, err := spdy.RoundTripperFor(kubernetesRestConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a SPDY round-tripper for the Kubernetes REST config")
	}
	dialerMethod := "POST"
	dialer := spdy.NewDialer(upgrader, &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}, dialerMethod, podProxyEndpointUrl)

	// Start forwarding ports asynchronously with reconnect logic.
	// The reconnect logic tries to reconnect after a working connection is lost.
	// The port forward process can be interrupted using the port forwarder stop channel.  There is no retry after that.
	go func() {
		retries := 0
		readyChannel := portforwardReadyChannel
		for {
			connection.portforwarder, err = portforward.NewOnAddresses(dialer, portForwardAddresses, portStrings, portforwardStopChannel, readyChannel, &portforwardStdOut, &portforwardStdErr)
			if err != nil {
				// Addresses or ports cannot be parsed so there is nothing else to try
				logrus.Errorf("An error occured parsing the port forwarder addresses or ports:\n%v", err)
				return
			} else {
				logrus.Debugf("Trying to forward ports for pod: %s", podProxyEndpointUrl.String())
				if err = connection.portforwarder.ForwardPorts(); err != nil {
					if err == portforward.ErrLostConnectionToPod {
						logrus.Infof("Lost connection to pod: %s", podProxyEndpointUrl.String())
						retries = 0
						// Copy the port forwarder assigned local ports so we re-use the same local ports when we reconnect
						ports, err := connection.portforwarder.GetPorts()
						if err != nil {
							logrus.Errorf("An error occured retrieving the local ports to remote ports mapping for our portforwarder:\n%v", err)
							return
						}
						portStrings = nil
						for _, port := range ports {
							portString := fmt.Sprintf("%v:%v", port.Local, port.Remote)
							portStrings = append(portStrings, portString)
						}
					} else {
						if retries == 0 {
							// Exit the retry logic if the first try to connect fails
							logrus.Errorf("Expected to be able to start forwarding local ports to remote ports, instead our portforwarder has returned a non-nil err:\n%v", err)
							return
						}
						logrus.Debugf("Error trying to forward ports:\n%v", err)
					}
				} else {
					// ForwardPorts() returns nil when we close the connection using the stop channel.
					// Do not try to reconnect.
					return
				}
				select {
				case <-stopChannel:
					return
				default:
				}
				time.Sleep(portForwardTimeBetweenRetries)
				retries += 1
				logrus.Debugf("Retrying (%d) connection to pod: %s", retries, podProxyEndpointUrl.String())
				readyChannel = make(chan struct{}, 1)
			}
		}
	}()

	// Wait for the portforwarder to be ready with timeout
	select {
	case <-portforwardReadyChannel:
	case <-time.After(portForwardTimeoutDuration):
		return nil, stacktrace.NewError("Expected Kubernetes portforwarder to open local ports to the pod exposed by the portforward api at URL '%v', instead the Kubernetes portforwarder timed out binding local ports", podProxyEndpointUrl)
	}
	// Get local forwarded ports
	forwardedPorts, err := connection.portforwarder.GetPorts()
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
		localPortSpec, err := port_spec.NewPortSpec(localPort, port_spec.TransportProtocol_TCP, emptyApplicationProtocol, noWait, "")
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to create port-spec describing local port '%v', instead a non-nil err was returned", localPort)
		}

		logrus.Infof("Forwarding requests to local port number '%v' to remote port with id '%v'", localPortSpec.GetNumber(), portSpecId)

		localPortSpecs[portSpecId] = localPortSpec
	}
	connection.localPorts = localPortSpecs

	return connection, nil
}

func (connection *gatewayConnectionToKurtosisImpl) Stop() {
	logrus.Infof("Closing connection to pod: %s", connection.urlString)
	close(connection.stopChannel)
	connection.portforwarder.Close()
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
	grpcConnection, err := grpc.Dial(localGrpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a GRPC client connection on address '%v', but a non-nil error was returned", localGrpcServerAddress)
	}

	return grpcConnection, nil
}

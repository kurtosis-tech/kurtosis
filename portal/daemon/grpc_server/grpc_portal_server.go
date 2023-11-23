package grpc_server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/portal/kurtosis_portal_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/portal/daemon/port_forward_manager"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	PortalServerGrpcPort = 9502
)

type GrpcPortalServer struct {
	sync.RWMutex

	portForwardManager *port_forward_manager.PortForwardManager
}

func NewPortalServer(manager *port_forward_manager.PortForwardManager) *GrpcPortalServer {
	return &GrpcPortalServer{
		RWMutex:            sync.RWMutex{},
		portForwardManager: manager,
	}
}

func (portalServer *GrpcPortalServer) Ping(ctx context.Context, ping *kurtosis_portal_rpc_api_bindings.PortalPing) (*kurtosis_portal_rpc_api_bindings.PortalPong, error) {
	err := portalServer.portForwardManager.Ping()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Portal Daemon is running but the Port Forward Manager failed to respond to the ping")
	}
	return &kurtosis_portal_rpc_api_bindings.PortalPong{}, nil
}

func (portalServer *GrpcPortalServer) ForwardUserServicePort(ctx context.Context, args *kurtosis_portal_rpc_api_bindings.ForwardUserServicePortArgs) (*kurtosis_portal_rpc_api_bindings.ForwardUserServicePortResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (portalServer *GrpcPortalServer) Close() error {
	portalServer.Lock()
	defer portalServer.Unlock()

	// TODO(omar): implement

	return nil
}

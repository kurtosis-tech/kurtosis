package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/portal/kurtosis_portal_rpc_api_bindings"
	"sync"
)

const (
	PortalServerGrpcPort = 9502
)

type PortalServer struct {
	sync.RWMutex
}

func (portalServer *PortalServer) Ping(ctx context.Context, ping *kurtosis_portal_rpc_api_bindings.PortalPing) (*kurtosis_portal_rpc_api_bindings.PortalPong, error) {
	//TODO implement me
	panic("implement me")
}

func (portalServer *PortalServer) ForwardUserServicePort(ctx context.Context, args *kurtosis_portal_rpc_api_bindings.ForwardUserServicePortArgs) (*kurtosis_portal_rpc_api_bindings.ForwardUserServicePortResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewPortalServer() *PortalServer {
	return &PortalServer{
		RWMutex: sync.RWMutex{},
	}
}

func (portalServer *PortalServer) Close() error {
	portalServer.Lock()
	defer portalServer.Unlock()

	// TODO(omar): implement

	return nil
}

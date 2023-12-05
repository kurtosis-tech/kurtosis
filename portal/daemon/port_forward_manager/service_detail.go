package port_forward_manager

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
)

type ServiceInterfaceDetail struct {
	enclaveServicePort EnclaveServicePort

	chiselServerUri string

	serviceIpAddress string

	servicePortSpec *services.PortSpec
}

func NewServiceDetail(esp EnclaveServicePort, chiselServerUri string, serviceIpAddress string, servicePortSpec *services.PortSpec) *ServiceInterfaceDetail {
	return &ServiceInterfaceDetail{
		esp,
		chiselServerUri,
		serviceIpAddress,
		servicePortSpec,
	}
}

func (sid ServiceInterfaceDetail) String() string {
	return fmt.Sprintf("%v: tunnel to '%v', remote at '%v:%d'", sid.enclaveServicePort, sid.chiselServerUri, sid.serviceIpAddress, sid.servicePortSpec.GetNumber())
}

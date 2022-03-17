package networking_sidecar

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

type NetworkingSidecarGUID string

type NetworkingSidecar struct {
	guid NetworkingSidecarGUID
	privateIpAddr net.IP
	enclaveId enclave.EnclaveID
}

func NewNetworkingSidecar(guid NetworkingSidecarGUID, privateIpAddr net.IP, enclaveId enclave.EnclaveID) *NetworkingSidecar {
	return &NetworkingSidecar{guid: guid, privateIpAddr: privateIpAddr, enclaveId: enclaveId}
}


func (sidecar *NetworkingSidecar) GetGuid() NetworkingSidecarGUID {
	return sidecar.guid
}

func (sidecar *NetworkingSidecar) GetPrivateIpAddr() net.IP {
	return sidecar.privateIpAddr
}

func (sidecar *NetworkingSidecar) GetEnclaveId() enclave.EnclaveID {
	return sidecar.enclaveId
}

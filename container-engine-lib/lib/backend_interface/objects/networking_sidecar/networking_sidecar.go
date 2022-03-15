package networking_sidecar

import "net"

type NetworkingSidecarGUID string

type NetworkingSidecar struct {
	guid NetworkingSidecarGUID
	privateIpAddr net.IP
}

func NewNetworkingSidecar(guid NetworkingSidecarGUID, privateIpAddr net.IP) *NetworkingSidecar {
	return &NetworkingSidecar{guid: guid, privateIpAddr: privateIpAddr}
}

func (sidecar *NetworkingSidecar) GetGuid() NetworkingSidecarGUID {
	return sidecar.guid
}

func (sidecar *NetworkingSidecar) GetPrivateIpAddr() net.IP {
	return sidecar.privateIpAddr
}

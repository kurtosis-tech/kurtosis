package service_network

import "net"

type ApiContainerInfo struct {
	ipAddress net.IP

	grpcPortNum uint16

	version string
}

func NewApiContainerInfo(
	ipAddress net.IP,
	grpcPortNum uint16,
	version string,
) *ApiContainerInfo {
	return &ApiContainerInfo{
		ipAddress:   ipAddress,
		grpcPortNum: grpcPortNum,
		version:     version,
	}
}

func (apic *ApiContainerInfo) GetIpAddress() net.IP {
	return apic.ipAddress
}

func (apic *ApiContainerInfo) GetGrpcPortNum() uint16 {
	return apic.grpcPortNum
}

func (apic *ApiContainerInfo) GetVersion() string {
	return apic.version
}

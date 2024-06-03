package service_network

import "net"

type ApiContainerInfo struct {
	ipAddress net.IP

	grpcPortNum uint16

	version string

	imageAuthor string
}

func NewApiContainerInfo(
	ipAddress net.IP,
	grpcPortNum uint16,
	version string,
	imageAuthor string,
) *ApiContainerInfo {
	return &ApiContainerInfo{
		ipAddress:   ipAddress,
		grpcPortNum: grpcPortNum,
		version:     version,
		imageAuthor: imageAuthor,
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

func (apic *ApiContainerInfo) GetImageAuthor() string {
	return apic.imageAuthor
}

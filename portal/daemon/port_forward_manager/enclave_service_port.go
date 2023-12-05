package port_forward_manager

type EnclaveServicePort struct {
	enclaveId string
	serviceId string
	portId    string
}

func NewEnclaveServicePort(enclaveId string, serviceId string, portId string) EnclaveServicePort {
	return EnclaveServicePort{
		enclaveId,
		serviceId,
		portId,
	}
}

func (esp EnclaveServicePort) String() string {
	return "(" + esp.enclaveId + ", " + esp.serviceId + ", " + esp.portId + ")"
}

func (esp EnclaveServicePort) EnclaveId() string {
	return esp.enclaveId
}

func (esp EnclaveServicePort) ServiceId() string {
	return esp.serviceId
}

func (esp EnclaveServicePort) PortId() string {
	return esp.portId
}

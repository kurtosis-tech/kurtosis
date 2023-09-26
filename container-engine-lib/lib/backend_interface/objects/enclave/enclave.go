package enclave

import "time"

type EnclaveUUID string

type Enclave struct {
	uuid                EnclaveUUID
	name                string
	status              EnclaveStatus
	creationTime        *time.Time
	isProductionEnclave bool
}

func NewEnclave(id EnclaveUUID, name string, status EnclaveStatus, creationTime *time.Time, productionMode bool) *Enclave {
	return &Enclave{uuid: id, name: name, status: status, creationTime: creationTime, isProductionEnclave: productionMode}
}

func (enclave *Enclave) GetUUID() EnclaveUUID {
	return enclave.uuid
}

func (enclave *Enclave) GetStatus() EnclaveStatus {
	return enclave.status
}

func (enclave *Enclave) GetCreationTime() *time.Time {
	return enclave.creationTime
}

func (enclave *Enclave) GetName() string {
	return enclave.name
}

func (enclave *Enclave) IsProductionEnclave() bool {
	return enclave.isProductionEnclave
}

package types

import (
	"time"
)

// Defines values for EnclaveStatus.
type EnclaveStatus string

const (
	EnclaveStatus_RUNNING EnclaveStatus = "RUNNING"
	EnclaveStatus_STOPPED EnclaveStatus = "STOPPED"
	EnclaveStatus_EMPTY   EnclaveStatus = "EMPTY"
)

// Defines values for ContainerStatus.
type ContainerStatus string

const (
	ContainerStatus_RUNNING     ContainerStatus = "RUNNING"
	ContainerStatus_STOPPED     ContainerStatus = "STOPPED"
	ContainerStatus_NONEXISTENT ContainerStatus = "NONEXISTENT"
)

// Defines values for EnclaveMode.
type EnclaveMode string

const (
	EnclaveMode_PRODUCTION EnclaveMode = "PRODUCTION"
	EnclaveMode_TEST       EnclaveMode = "TEST"
)

// CreateEnclaveArgs defines model for CreateEnclaveArgs.
type CreateEnclaveArgs struct {
	ApiContainerLogLevel   string
	ApiContainerVersionTag string
	EnclaveName            string
	Mode                   *EnclaveMode
}

// EnclaveAPIContainerHostMachineInfo defines model for EnclaveAPIContainerHostMachineInfo.
type EnclaveAPIContainerHostMachineInfo struct {
	GrpcPortOnHostMachine uint32
	IpOnHostMachine       string
}

// EnclaveAPIContainerInfo defines model for EnclaveAPIContainerInfo.
type EnclaveAPIContainerInfo struct {
	BridgeIpAddress       string
	ContainerId           string
	GrpcPortInsideEnclave uint32
	IpInsideEnclave       string
}

// EnclaveIdentifiers defines model for EnclaveIdentifiers.
type EnclaveIdentifiers struct {
	EnclaveUuid   string
	Name          string
	ShortenedUuid string
}

// EnclaveInfo defines model for EnclaveInfo.
type EnclaveInfo struct {
	ApiContainerHostMachineInfo *EnclaveAPIContainerHostMachineInfo
	ApiContainerInfo            *EnclaveAPIContainerInfo
	ApiContainerStatus          ContainerStatus
	EnclaveStatus               EnclaveStatus
	CreationTime                Timestamp
	EnclaveUuid                 string
	Mode                        EnclaveMode
	Name                        string
	ShortenedUuid               string
}

// EnclaveNameAndUuid defines model for EnclaveNameAndUuid.
type EnclaveNameAndUuid struct {
	Name string
	Uuid string
}

type GetEnclavesResponse struct {
	EnclaveInfo *map[string]EnclaveInfo
}

// GetEngineInfoResponse defines model for GetEngineInfoResponse.
type EngineVersion struct {
	EngineVersion string
}

// GetServiceLogsArgs defines model for GetServiceLogsArgs.
type GetServiceLogsArgs struct {
	ConjunctiveFilters *[]LogLineFilter
	FollowLogs         bool
	NumLogLines        int
	ReturnAllLogs      bool
	ServiceUuidSet     *[]string
}

// GetServiceLogsResponse defines model for GetServiceLogsResponse.
type GetServiceLogsResponse struct {
	NotFoundServiceUuidSet   *[]string
	ServiceLogsByServiceUuid *map[string]LogLine
}

// LogLine defines model for LogLine.
type LogLine struct {
	Line      *[]string
	Timestamp Timestamp
}

// LogLineFilter defines model for LogLineFilter.
type LogLineFilter struct {
	Operator    *int
	TextPattern *string
}

// Timestamp defines model for Timestamp.
type Timestamp = time.Time

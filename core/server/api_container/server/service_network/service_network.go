package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"io"
	"net/http"
)

// ServiceNetwork handles the state of the enclave
//
// Regenerate mock with the following command from core/server directory:
// mockery -r --name=ServiceNetwork --filename=mock_service_network.go --structname=MockServiceNetwork --with-expecter --inpackage
type ServiceNetwork interface {
	AddService(
		ctx context.Context,
		serviceName service.ServiceName,
		serviceConfig *service.ServiceConfig,
	) (
		*service.Service,
		error,
	)

	AddServices(
		ctx context.Context,
		serviceConfigs map[service.ServiceName]*service.ServiceConfig,
		batchSize int,
	) (
		map[service.ServiceName]*service.Service,
		map[service.ServiceName]error,
		error,
	)

	UpdateService(
		ctx context.Context,
		serviceName service.ServiceName,
		updateServiceConfig *service.ServiceConfig,
	) (
		*service.Service,
		error,
	)

	UpdateServices(
		ctx context.Context,
		updateServiceConfigs map[service.ServiceName]*service.ServiceConfig,
		batchSize int,
	) (
		map[service.ServiceName]*service.Service,
		map[service.ServiceName]error,
		error,
	)

	RemoveService(ctx context.Context, serviceIdentifier string) (service.ServiceUUID, error)

	StartService(ctx context.Context, serviceIdentifier string) error

	StartServices(
		ctx context.Context,
		serviceIdentifiers []string,
	) (
		map[service.ServiceUUID]bool,
		map[service.ServiceUUID]error,
		error,
	)

	StopService(ctx context.Context, serviceIdentifier string) error

	StopServices(
		ctx context.Context,
		serviceIdentifiers []string,
	) (
		map[service.ServiceUUID]bool,
		map[service.ServiceUUID]error,
		error,
	)

	RunExec(ctx context.Context, serviceIdentifier string, userServiceCommand []string) (*exec_result.ExecResult, error)

	RunExecs(
		ctx context.Context,
		userServiceCommands map[string][]string,
	) (
		map[service.ServiceUUID]*exec_result.ExecResult,
		map[service.ServiceUUID]error,
		error,
	)

	HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string) (*http.Response, error)

	GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error)

	CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	GetServiceNames() map[service.ServiceName]bool

	GetExistingAndHistoricalServiceIdentifiers() []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers

	GetServiceRegistration(serviceName service.ServiceName) (*service.ServiceRegistration, bool)

	RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*render_templates.TemplateData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	UploadFilesArtifact(data io.Reader, contentMd5 []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	GetFilesArtifactMd5(artifactName string) (enclave_data_directory.FilesArtifactUUID, []byte, bool, error)

	UpdateFilesArtifact(fileArtifactUuid enclave_data_directory.FilesArtifactUUID, updatedContent io.Reader, contentMd5 []byte) error

	GetUniqueNameForFileArtifact() (string, error)

	GetApiContainerInfo() *ApiContainerInfo
}

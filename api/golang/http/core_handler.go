package http

import (
	api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_http_api_bindings"
	"github.com/labstack/echo/v4"
)

type EngineRuntime struct {
}

type Error struct {
}

func (error Error) Error() string {
	return "Not Implemented :("
}

func (engine *EngineRuntime) GetArtifacts(ctx echo.Context) error {
	return Error{}
}

// (PUT /artifacts/local-file)
func (engine *EngineRuntime) PutArtifactsLocalFile(ctx echo.Context) error {
	return Error{}
}

// (PUT /artifacts/remote-file)
func (engine *EngineRuntime) PutArtifactsRemoteFile(ctx echo.Context) error {
	return Error{}
}

// (PUT /artifacts/services/{service_identifier})
func (engine *EngineRuntime) PutArtifactsServicesServiceIdentifier(ctx echo.Context, serviceIdentifier string) error {
	return Error{}
}

// (GET /artifacts/{artifact_identifier})
func (engine *EngineRuntime) GetArtifactsArtifactIdentifier(ctx echo.Context, artifactIdentifier string) error {
	return Error{}
}

// (GET /artifacts/{artifact_identifier}/download)
func (engine *EngineRuntime) GetArtifactsArtifactIdentifierDownload(ctx echo.Context, artifactIdentifier string) error {
	return Error{}
}

// (GET /services)
func (engine *EngineRuntime) GetServices(ctx echo.Context) error {
	return Error{}
}

// (POST /services/connection)
func (engine *EngineRuntime) PostServicesConnection(ctx echo.Context) error {
	return Error{}
}

// (GET /services/{service_identifier})
func (engine *EngineRuntime) GetServicesServiceIdentifier(ctx echo.Context, serviceIdentifier string, params api.GetServicesServiceIdentifierParams) error {
	return Error{}
}

// (POST /services/{service_identifier}/command)
func (engine *EngineRuntime) PostServicesServiceIdentifierCommand(ctx echo.Context, serviceIdentifier string) error {
	return Error{}
}

// (POST /services/{service_identifier}/endpoints/{port_number}/availability)
func (engine *EngineRuntime) PostServicesServiceIdentifierEndpointsPortNumberAvailability(ctx echo.Context, serviceIdentifier string, portNumber int) error {
	return Error{}
}

// (GET /starlark)
func (engine *EngineRuntime) GetStarlark(ctx echo.Context) error {
	return Error{}
}

// (PUT /starlark/packages)
func (engine *EngineRuntime) PutStarlarkPackages(ctx echo.Context) error {
	return Error{}
}

// (POST /starlark/packages/{package_id})
func (engine *EngineRuntime) PostStarlarkPackagesPackageId(ctx echo.Context, packageId string) error {
	return Error{}
}

// (POST /starlark/scripts)
func (engine *EngineRuntime) PostStarlarkScripts(ctx echo.Context) error {
	return Error{}
}

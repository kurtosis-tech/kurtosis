// Package websocket_api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package websocket_api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get enclave's services logs
	// (GET /enclaves/{enclave_identifier}/logs)
	GetEnclavesEnclaveIdentifierLogs(ctx echo.Context, enclaveIdentifier EnclaveIdentifier, params GetEnclavesEnclaveIdentifierLogsParams) error
	// Get service logs
	// (GET /enclaves/{enclave_identifier}/services/{service_identifier}/logs)
	GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs(ctx echo.Context, enclaveIdentifier EnclaveIdentifier, serviceIdentifier ServiceIdentifier, params GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogsParams) error
	// Get Starlark execution logs
	// (GET /starlark/executions/{starlark_execution_uuid}/logs)
	GetStarlarkExecutionsStarlarkExecutionUuidLogs(ctx echo.Context, starlarkExecutionUuid StarlarkExecutionUuid) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetEnclavesEnclaveIdentifierLogs converts echo context to params.
func (w *ServerInterfaceWrapper) GetEnclavesEnclaveIdentifierLogs(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "enclave_identifier" -------------
	var enclaveIdentifier EnclaveIdentifier

	err = runtime.BindStyledParameterWithLocation("simple", false, "enclave_identifier", runtime.ParamLocationPath, ctx.Param("enclave_identifier"), &enclaveIdentifier)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter enclave_identifier: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetEnclavesEnclaveIdentifierLogsParams
	// ------------- Required query parameter "service_uuid_set" -------------

	err = runtime.BindQueryParameter("form", true, true, "service_uuid_set", ctx.QueryParams(), &params.ServiceUuidSet)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter service_uuid_set: %s", err))
	}

	// ------------- Optional query parameter "follow_logs" -------------

	err = runtime.BindQueryParameter("form", true, false, "follow_logs", ctx.QueryParams(), &params.FollowLogs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter follow_logs: %s", err))
	}

	// ------------- Optional query parameter "conjunctive_filters" -------------

	err = runtime.BindQueryParameter("form", true, false, "conjunctive_filters", ctx.QueryParams(), &params.ConjunctiveFilters)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter conjunctive_filters: %s", err))
	}

	// ------------- Optional query parameter "return_all_logs" -------------

	err = runtime.BindQueryParameter("form", true, false, "return_all_logs", ctx.QueryParams(), &params.ReturnAllLogs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter return_all_logs: %s", err))
	}

	// ------------- Optional query parameter "num_log_lines" -------------

	err = runtime.BindQueryParameter("form", true, false, "num_log_lines", ctx.QueryParams(), &params.NumLogLines)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter num_log_lines: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetEnclavesEnclaveIdentifierLogs(ctx, enclaveIdentifier, params)
	return err
}

// GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs converts echo context to params.
func (w *ServerInterfaceWrapper) GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "enclave_identifier" -------------
	var enclaveIdentifier EnclaveIdentifier

	err = runtime.BindStyledParameterWithLocation("simple", false, "enclave_identifier", runtime.ParamLocationPath, ctx.Param("enclave_identifier"), &enclaveIdentifier)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter enclave_identifier: %s", err))
	}

	// ------------- Path parameter "service_identifier" -------------
	var serviceIdentifier ServiceIdentifier

	err = runtime.BindStyledParameterWithLocation("simple", false, "service_identifier", runtime.ParamLocationPath, ctx.Param("service_identifier"), &serviceIdentifier)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter service_identifier: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogsParams
	// ------------- Optional query parameter "follow_logs" -------------

	err = runtime.BindQueryParameter("form", true, false, "follow_logs", ctx.QueryParams(), &params.FollowLogs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter follow_logs: %s", err))
	}

	// ------------- Optional query parameter "conjunctive_filters" -------------

	err = runtime.BindQueryParameter("form", true, false, "conjunctive_filters", ctx.QueryParams(), &params.ConjunctiveFilters)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter conjunctive_filters: %s", err))
	}

	// ------------- Optional query parameter "return_all_logs" -------------

	err = runtime.BindQueryParameter("form", true, false, "return_all_logs", ctx.QueryParams(), &params.ReturnAllLogs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter return_all_logs: %s", err))
	}

	// ------------- Optional query parameter "num_log_lines" -------------

	err = runtime.BindQueryParameter("form", true, false, "num_log_lines", ctx.QueryParams(), &params.NumLogLines)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter num_log_lines: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs(ctx, enclaveIdentifier, serviceIdentifier, params)
	return err
}

// GetStarlarkExecutionsStarlarkExecutionUuidLogs converts echo context to params.
func (w *ServerInterfaceWrapper) GetStarlarkExecutionsStarlarkExecutionUuidLogs(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "starlark_execution_uuid" -------------
	var starlarkExecutionUuid StarlarkExecutionUuid

	err = runtime.BindStyledParameterWithLocation("simple", false, "starlark_execution_uuid", runtime.ParamLocationPath, ctx.Param("starlark_execution_uuid"), &starlarkExecutionUuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter starlark_execution_uuid: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetStarlarkExecutionsStarlarkExecutionUuidLogs(ctx, starlarkExecutionUuid)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/enclaves/:enclave_identifier/logs", wrapper.GetEnclavesEnclaveIdentifierLogs)
	router.GET(baseURL+"/enclaves/:enclave_identifier/services/:service_identifier/logs", wrapper.GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs)
	router.GET(baseURL+"/starlark/executions/:starlark_execution_uuid/logs", wrapper.GetStarlarkExecutionsStarlarkExecutionUuidLogs)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"ogiZ2GLpWnrR+rNb+bD/TwAAAP//53BKX7omAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

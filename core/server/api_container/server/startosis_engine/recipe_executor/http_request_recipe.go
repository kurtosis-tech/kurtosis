package recipe_executor

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"io"
)

type HttpRequestRecipe struct {
	serviceId   service.ServiceID
	portId      string
	contentType string
	endpoint    string
	method      string
	body        string
}

type HttpRequestRuntimeValue struct {
	body string
	code int
}

func NewPostHttpRequestRecipe(serviceId service.ServiceID, portId string, contentType string, endpoint string, body string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      "POST",
		contentType: contentType,
		endpoint:    endpoint,
		body:        body,
	}
}

func NewGetHttpRequestRecipe(serviceId service.ServiceID, portId string, endpoint string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      "GET",
		contentType: "",
		endpoint:    endpoint,
		body:        "",
	}
}

func (recipe *HttpRequestRecipe) Execute(ctx context.Context, serviceNetwork service_network.ServiceNetwork) (map[string]string, error) {
	response, err := serviceNetwork.HttpRequestService(
		ctx,
		recipe.serviceId,
		recipe.portId,
		recipe.method,
		recipe.contentType,
		recipe.endpoint,
		recipe.body,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe")
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logrus.Errorf("An error occurred when closing response body: %v", err)
		}
	}()
	body, err := io.ReadAll(response.Body)
	logrus.Debugf("Got response '%v'", string(body))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when reading HTTP response body")
	}
	return map[string]string{
		"body": string(body),
		"code": fmt.Sprint(response.StatusCode),
	}, nil
}

func CreateStarlarkDictFromHttpRequestRuntimeValue(bodyMagicString starlark.String, codeMagicString starlark.String) *starlark.Dict {
	dict := &starlark.Dict{}
	dict.SetKey(starlark.String("body"), bodyMagicString)
	dict.SetKey(starlark.String("code"), codeMagicString)
	return dict
}

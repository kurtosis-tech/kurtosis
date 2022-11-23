package recipe_executor

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"io"
)

const (
	postMethod        = "POST"
	getMethod         = "GET"
	emptyBody         = ""
	unusedContentType = ""
)

type HttpRequestRecipe struct {
	serviceId   service.ServiceID
	portId      string
	contentType string
	endpoint    string
	method      string
	body        string
}

func NewPostHttpRequestRecipe(serviceId service.ServiceID, portId string, contentType string, endpoint string, body string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      postMethod,
		contentType: contentType,
		endpoint:    endpoint,
		body:        body,
	}
}

func NewGetHttpRequestRecipe(serviceId service.ServiceID, portId string, endpoint string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      getMethod,
		contentType: unusedContentType,
		endpoint:    endpoint,
		body:        emptyBody,
	}
}

func (recipe *HttpRequestRecipe) Execute(ctx context.Context, serviceNetwork service_network.ServiceNetwork) (map[string]starlark.Comparable, error) {
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
	return map[string]starlark.Comparable{
		"body": starlark.String(body),
		"code": starlark.MakeInt(response.StatusCode),
	}, nil
}

func CreateStarlarkStructFromHttpRequestRuntimeValue(bodyMagicString starlark.String, codeMagicString starlark.String) *starlarkstruct.Struct {
	return starlarkstruct.FromKeywords(
		starlarkstruct.Default,
		[]starlark.Tuple{
			{starlark.String("body"), bodyMagicString},
			{starlark.String("code"), codeMagicString},
		},
	)
}

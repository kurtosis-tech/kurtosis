package centralized_logs

import (
	"io"
	"net/http"
	"strings"
)

type mockedHttpClient struct {
	request *http.Request
}

func NewMockedHttpClient() *mockedHttpClient {
	return &mockedHttpClient{}
}

func (client *mockedHttpClient) Do(request *http.Request) (*http.Response, error) {

	client.request = request

	response := &http.Response{}

	response.StatusCode = http.StatusOK

	mockedResponseBodyReader := strings.NewReader(mockedResponseBodyStr)
	mockedResponseBodyReadCloser := io.NopCloser(mockedResponseBodyReader)

	response.Body = mockedResponseBodyReadCloser

	return response, nil
}

func (client *mockedHttpClient) GetRequest() *http.Request {
	return client.request
}

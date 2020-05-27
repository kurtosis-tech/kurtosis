package testsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gmarchetti/kurtosis/ava_commons/networks"
	"github.com/gmarchetti/kurtosis/commons/testsuite"
	"io/ioutil"
	"net/http"
)

type SingleNodeGeckoNetworkBasicTest struct {}
func (s SingleNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(networks.SingleNodeGeckoNetwork)
	httpSocket := castedNetwork.GetNode().GetJsonRpcSocket()

	requestBody, err := json.Marshal(map[string]string{
		"jsonrpc": "2.0",
		"id": "1",
		"method": "admin.peers",
	})
	if err != nil {
		context.Fatal(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%v:%v/ext/admin", httpSocket.GetIpAddr(), httpSocket.GetPort()),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		context.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		context.Fatal(err)
	}

	// TODO parse the response as JSON and assert that we get the expected number of peers
	println(string(body))
}



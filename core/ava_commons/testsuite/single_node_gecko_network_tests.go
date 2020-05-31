package testsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/ava_commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const (
)

type SingleNodeGeckoNetworkBasicTest struct {}
func (test SingleNodeGeckoNetworkBasicTest) Run(network interface{}, context testsuite.TestContext) {
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
		fmt.Sprintf("http://%v:%v/ext/admin", httpSocket.GetIpAddr(), httpSocket.GetPort()),
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

type SingleNodeNetworkGetValidatorsTest struct{}
func (test SingleNodeNetworkGetValidatorsTest) Run(network interface{}, context testsuite.TestContext) {
	castedNetwork := network.(networks.SingleNodeGeckoNetwork)

	// TODO Move these into a better location
	RPC_BODY := `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`
	RETRIES := 5
	RETRY_WAIT_SECONDS := 5*time.Second

	// Run RPC Test on PChain.
	var jsonStr = []byte(RPC_BODY)
	var jsonBuffer = bytes.NewBuffer(jsonStr)
	logrus.Infof("Test request as string: %s", jsonBuffer.String())

	var validatorList ValidatorList
	jsonRpcSocket := castedNetwork.GetNode().GetJsonRpcSocket()
	endpoint := fmt.Sprintf("http://%v:%v/%v", jsonRpcSocket.GetIpAddr(), jsonRpcSocket.GetPort().Int(), GetPChainEndpoint())
	for i := 0; i < RETRIES; i++ {
		resp, err := http.Post(endpoint, "application/json", jsonBuffer)
		if err != nil {
			logrus.Infof("Attempted connection...: %s", err.Error())
			logrus.Infof("Could not connect on attempt %d, retrying...", i+1)
			time.Sleep(RETRY_WAIT_SECONDS)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatalln(err)
		}

		var validatorResponse ValidatorResponse
		json.Unmarshal(body, &validatorResponse)

		validatorList = validatorResponse.Result["validators"]
		if len(validatorList) > 0 {
			logrus.Infof("Found validators!")
			break
		}
	}
	for _, validator := range validatorList {
		logrus.Infof("Validator id: %s", validator.Id)
	}
	if len(validatorList) < 1 {
		logrus.Infof("Failed to find a single validator.")
	}
}


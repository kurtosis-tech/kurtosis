package fluentbit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net/http"
	"text/template"
	"time"
)

const (
	localhostStr    = "localhost"
	httpProtocolStr = "http"

	waitForAvailabilityInitialDelayMilliseconds = 100
	waitForAvailabilityMaxRetries               = 20
	waitForAvailabilityRetriesDelayMilliseconds = 50
)

type Fluentbit struct {
	config *Config
}

func NewFluentbit(config *Config) *Fluentbit {
	return &Fluentbit{config: config}
}

func (fluent *Fluentbit) GetPrivateTcpPortSpec() (*port_spec.PortSpec, error) {
	privateTcpPortSpec, err := port_spec.NewPortSpec(tcpPortNumber, tcpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit's private TCP port spec object using number '%v' and protocol '%v'",
			tcpPortNumber,
			tcpPortProtocol,
		)
	}
	return privateTcpPortSpec, nil
}

func (fluent *Fluentbit) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(httpPortNumber, httpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit's private HTTP port spec object using number '%v' and protocol '%v'",
			httpPortNumber,
			httpPortProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (fluent *Fluentbit) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	volumeName string,
	networkId string,
	dockerManager *docker_manager.DockerManager,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	privateTcpPortSpec, err := fluent.GetPrivateTcpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Fluentbit's private TCP port spec")
	}

	privateTcpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateTcpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private TCP port spec to a Docker port")
	}

	privateHttpPortSpec, err := fluent.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Fluentbit's private HTTP port spec")
	}

	privateHttpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateHttpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private HTTP port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateTcpDockerPort:  docker_manager.NewNoPublishingSpec(),
		privateHttpDockerPort: docker_manager.NewManualPublishingSpec(httpPortNumber),
	}

	volumeMounts := map[string]string{
		volumeName: configDirpathInContainer,
	}

	if err := fluent.runConfigurator(context.Background(), networkId, volumeMounts, dockerManager); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the Fluenbit configurator in network ID '%v' and with volume mounts '%+v'",
			networkId,
			volumeMounts,
		)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithUsedPorts(
		usedPorts,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	return createAndStartArgs, nil
}

func (fluent *Fluentbit) WaitForAvailability() error {
	return waitForEndpointAvailability(
		localhostStr,
		httpPortNumber,
		healthCheckEndpointPath,
		waitForAvailabilityInitialDelayMilliseconds,
		waitForAvailabilityMaxRetries,
		waitForAvailabilityRetriesDelayMilliseconds,
	)
}

func (fluent *Fluentbit) getConfigFileContent() (string, error) {

	template, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Fluenbit config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	template.Execute(templateStrBuffer, fluent.config)

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}

func waitForEndpointAvailability(
	host string,
	port uint16,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
) error {

	var err error

	url := fmt.Sprintf("%v://%v:%v/%v", httpProtocolStr, host, port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		_, err = makeHttpRequest(url)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	return nil
}

func makeHttpRequest(url string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	resp, err = http.Get(url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v' ", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

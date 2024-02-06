package object_attributes_provider

import (
	"net"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
)

const (
	enclaveUuid = "65d2fb6d673249b8b4a91a2f4ae616de"
)

var (
	portWaitForTest = port_spec.NewWait(5 * time.Second)
)

func TestForUserServiceContainer(t *testing.T) {
	objAttrsProvider := GetDockerObjectAttributesProvider()
	enclaveObjAttrsProvider, err := objAttrsProvider.ForEnclave(enclaveUuid)
	require.NoError(t, err, "An unexpected error occurred getting the enclave object attributes provider")

	serviceName := service.ServiceName("nginx")
	serviceUuid := service.ServiceUUID("3771c85af16a40a18201acf4b4b5ad28")
	privateIpAddr := net.IP("1.2.3.4")
	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, "", portWaitForTest)
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")
	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	port2ApplicationProtocol := consts.HttpApplicationProtocol
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, port2ApplicationProtocol, portWaitForTest)
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")
	privatePorts := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}
	userLabels := map[string]string{}
	containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(
		serviceName,
		serviceUuid,
		privateIpAddr,
		privatePorts,
		userLabels,
	)
	require.NoError(t, err, "An unexpected error occurred getting the container attributes")
	objName := containerAttrs.GetName()
	require.Equal(t, objName.GetString(), "nginx--3771c85af16a40a18201acf4b4b5ad28")
	objLabels := containerAttrs.GetLabels()
	for labelKey, labelValue := range objLabels {
		switch labelKey.GetString() {
		case docker_label_key.AppIDDockerLabelKey.GetString():
			require.Equal(t, labelValue.GetString(), "kurtosis")
		case docker_label_key.ContainerTypeDockerLabelKey.GetString():
			require.Equal(t, labelValue.GetString(), "user-service")
		case docker_label_key.EnclaveUUIDDockerLabelKey.GetString():
			require.Equal(t, labelValue.GetString(), "65d2fb6d673249b8b4a91a2f4ae616de")
		case "traefik.enable":
			require.Equal(t, labelValue.GetString(), "true")
		case "traefik.http.routers.65d2fb6d6732-3771c85af16a-23.rule":
			require.Fail(t, "A traefik label for port 23 should not be present")
		case "traefik.http.routers.65d2fb6d6732-3771c85af16a-45.rule":
			require.Equal(t, labelValue.GetString(), "Host(`45-3771c85af16a-65d2fb6d6732`)")
		case "traefik.http.routers.65d2fb6d6732-3771c85af16a-45.service":
			require.Equal(t, labelValue.GetString(), "65d2fb6d6732-3771c85af16a-45")
		case "traefik.http.services.65d2fb6d6732-3771c85af16a-45.loadbalancer.server.port":
			require.Equal(t, labelValue.GetString(), "45")
		default:
			break
		}
	}
}

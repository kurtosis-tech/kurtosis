package service_identifiers

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

func getServiceIdentifiersForTest() []*serviceIdentifier {
	serviceUuid1 := service.ServiceUUID("33faa0b925e142b9b2ef6ebcd5cc7266")
	serviceName1 := service.ServiceName("service-for-test-1")
	serviceIdentifier1 := NewServiceIdentifier(serviceUuid1, serviceName1)

	serviceUuid2 := service.ServiceUUID("44dea0b925e142b9b2ef6aeba5cc6589")
	serviceName2 := service.ServiceName("service-for-test-2")
	serviceIdentifier2 := NewServiceIdentifier(serviceUuid2, serviceName2)

	return []*serviceIdentifier{serviceIdentifier1, serviceIdentifier2}
}

package service_identifiers

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	uuidKey             = "uuid"
	shortenedUuidStrKey = "shortened_uuid_str"
	nameKey             = "name"
)

type serviceIdentifier struct {
	uuid             service.ServiceUUID
	shortenedUuidStr string
	name             service.ServiceName
}

func NewServiceIdentifier(uuid service.ServiceUUID, name service.ServiceName) *serviceIdentifier {
	uuidStr := string(uuid)
	shortenedUuidStr := uuid_generator.ShortenedUUIDString(uuidStr)

	return &serviceIdentifier{uuid: uuid, shortenedUuidStr: shortenedUuidStr, name: name}
}

func (serviceIdentifier *serviceIdentifier) GetUuid() service.ServiceUUID {
	return serviceIdentifier.uuid
}

func (serviceIdentifier *serviceIdentifier) GetShortenedUUIDStr() string {
	return serviceIdentifier.shortenedUuidStr
}

func (serviceIdentifier *serviceIdentifier) GetName() service.ServiceName {
	return serviceIdentifier.name
}

func (serviceIdentifier *serviceIdentifier) MarshalJSON() ([]byte, error) {

	data := map[string]string{
		uuidKey:             string(serviceIdentifier.uuid),
		shortenedUuidStrKey: serviceIdentifier.shortenedUuidStr,
		nameKey:             string(serviceIdentifier.name),
	}

	return json.Marshal(data)
}

func (serviceIdentifier *serviceIdentifier) UnmarshalJSON(data []byte) error {

	unmarshalledMapPtr := &map[string]string{}

	if err := json.Unmarshal(data, unmarshalledMapPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling map")
	}

	unmarshalledMap := *unmarshalledMapPtr

	uuidStr, found := unmarshalledMap[uuidKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", uuidKey, unmarshalledMap)
	}
	uuid := service.ServiceUUID(uuidStr)

	shortenedUuidStr, found := unmarshalledMap[shortenedUuidStrKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", shortenedUuidStrKey, unmarshalledMap)
	}

	nameStr, found := unmarshalledMap[nameKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", nameKey, unmarshalledMap)
	}
	name := service.ServiceName(nameStr)

	serviceIdentifier.uuid = uuid
	serviceIdentifier.shortenedUuidStr = shortenedUuidStr
	serviceIdentifier.name = name
	return nil
}

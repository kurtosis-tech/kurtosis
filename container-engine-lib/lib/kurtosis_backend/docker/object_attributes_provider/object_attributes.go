package schema

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/stacktrace"
)

// Encapsulates the attributes that a Docker object in the Kurtosis ecosystem can have
type DockerObjectAttributes interface {
	GetName() string
	GetLabels() map[string]string
}

// Private so this can't be instantiated
type dockerObjectAttributesImpl struct {
	name         string
	customLabels map[string]string
}

func newDockerObjectAttributesImpl(name *docker.DockerObjectName, customLabels map[*docker.DockerLabelKey]*docker.DockerLabelValue) (*dockerObjectAttributesImpl, error) {
	for key, value := range customLabels {
		if err := validateLabelKey(key); err != nil {
			return nil, stacktrace.Propagate(err, "The label key '%s' for value '%s' is invalid", key, value)
		}
		if err := validateLabelValue(value); err != nil {
			return nil, stacktrace.Propagate(err, "The label value '%s' for key '%s' is invalid", value, key)
		}
	}

	return &objectAttributesImpl{name: name, customLabels: customLabels}, nil
}

func (attrs *dockerObjectAttributesImpl) GetName() string {
	return attrs.name
}

func (attrs *dockerObjectAttributesImpl) GetLabels() map[*docker.DockerLabelKey]*docker.DockerLabelValue {
	result := map[string]string{}
	for key, value := range attrs.customLabels {
		result[key] = value
	}
	// NOTE: If a custom label collides with a base label, the base label wins
	for key, value := range forever_constants.ForeverLabels {
		result[key] = value
	}
	return result
}

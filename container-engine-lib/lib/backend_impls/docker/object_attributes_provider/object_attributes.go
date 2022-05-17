package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
)

// Labels that get attached to EVERY Kurtosis object
var globalLabels = map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
	label_key_consts.AppIDLabelKey: label_value_consts.AppIDKubernetesLabelValue,
	// TODO container engine lib version??
}

// Encapsulates the attributes that a Docker object in the Kurtosis ecosystem can have
type DockerObjectAttributes interface {
	GetName() *docker_object_name.DockerObjectName
	GetLabels() map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue
}

// Private so this can't be instantiated
type dockerObjectAttributesImpl struct {
	name         *docker_object_name.DockerObjectName
	customLabels map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue
}

func newDockerObjectAttributesImpl(name *docker_object_name.DockerObjectName, customLabels map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue) (*dockerObjectAttributesImpl, error) {
	globalLabelsStrs := map[string]string{}
	for globalKey, globalValue := range globalLabels {
		globalLabelsStrs[globalKey.GetString()] = globalValue.GetString()
	}
	for customKey, customValue := range customLabels {
		if _, found := globalLabelsStrs[customKey.GetString()]; found {
			return nil, stacktrace.NewError("Custom label with key '%v' and value '%v' collides with global label with the same key", customKey.GetString(), customValue.GetString())
		}
	}

	return &dockerObjectAttributesImpl{
		name:         name,
		customLabels: customLabels,
	}, nil
}

func (attrs *dockerObjectAttributesImpl) GetName() *docker_object_name.DockerObjectName {
	return attrs.name
}

func (attrs *dockerObjectAttributesImpl) GetLabels() map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue {
	result := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{}
	for key, value := range attrs.customLabels {
		result[key] = value
	}
	// We're guaranteed that the global label string keys won't collide with the custom labels due to the validation
	// we do at construction time
	for key, value := range globalLabels {
		result[key] = value
	}
	return result
}

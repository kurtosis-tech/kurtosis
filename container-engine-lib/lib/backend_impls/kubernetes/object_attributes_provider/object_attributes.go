package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
)

// Labels that get attached to EVERY Kurtosis object
var globalLabels = map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
	label_key_consts.AppIDKubernetesLabelKey: label_value_consts.AppIDLabelValue,
	// TODO container engine lib version??
}

// Encapsulates the attributes that a Kubernetes object in the Kurtosis ecosystem can have
type KubernetesObjectAttributes interface {
	GetName() *kubernetes_object_name.KubernetesObjectName
	GetLabels() map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue
	GetAnnotations() map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue
}

// Private so this can't be instantiated
type kubernetesObjectAttributesImpl struct {
	name              *kubernetes_object_name.KubernetesObjectName
	customLabels      map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue
	customAnnotations map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue
}

func newKubernetesObjectAttributesImpl(name *kubernetes_object_name.KubernetesObjectName, customLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, customAnnotations map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue) (*kubernetesObjectAttributesImpl, error) {
	globalLabelsStrs := map[string]string{}
	for globalKey, globalValue := range globalLabels {
		globalLabelsStrs[globalKey.GetString()] = globalValue.GetString()
	}
	for customKey, customValue := range customLabels {
		if _, found := globalLabelsStrs[customKey.GetString()]; found {
			return nil, stacktrace.NewError("Custom label with key '%v' and value '%v' collides with global label with the same key", customKey.GetString(), customValue.GetString())
		}
	}

	return &kubernetesObjectAttributesImpl{
		name:              name,
		customLabels:      customLabels,
		customAnnotations: customAnnotations,
	}, nil
}

func (attrs *kubernetesObjectAttributesImpl) GetName() *kubernetes_object_name.KubernetesObjectName {
	return attrs.name
}

func (attrs *kubernetesObjectAttributesImpl) GetLabels() map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue {
	result := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{}
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

func (attrs *kubernetesObjectAttributesImpl) GetAnnotations() map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue {
	return attrs.customAnnotations
}

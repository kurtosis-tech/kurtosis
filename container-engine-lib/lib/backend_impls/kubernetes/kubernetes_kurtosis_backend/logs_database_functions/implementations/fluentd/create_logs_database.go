package fluentd

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

const (
	kurtosisLogsDatabaseContainerName = "kurtosis-logs-database-container"
)

type fluentdContainer struct{}

func NewFluentdContainer() *fluentdContainer {
	return &fluentdContainer{}
}

func (fluentd *fluentdContainer) CreateAndStart(
	ctx context.Context,
	engineNamespace string,
	serviceAccountName string,
	grpcPortNum uint16,
	envVars map[string]string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
) (string, uint16, error) {
	logsDatabaseGuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := logs_database.LogsDatabaseGUID(logsDatabaseGuid)
	logsDatabaseAttributesProvider := objAttrsProvider.ForLogsDatabase(engineGuid)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.KurtosisServersTransportProtocol, consts.HttpApplicationProtocol, nil)
	if err != nil {
		return "", 0, stacktrace.Propagate(
			err,
			"An error occurred creating the logs database's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.KurtosisServersTransportProtocol.String(),
		)
	}
	privatePortSpecs := map[string]*port_spec.PortSpec{
		consts.KurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
	}

	logsDatabasePod, logsDatabasePodLabels, err := createLogsDatabasePod(ctx, engineNamespace, serviceAccountName, envVars, privatePortSpecs, kubernetesManager, logsDatabaseAttributesProvider)
	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred creating the logs database pod")
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err = kubernetesManager.RemovePod(ctx, logsDatabasePod); err != nil {
				logrus.Errorf("Creating the logs database didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", logsDatabasePod.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", logsDatabasePod.Name)
			}
		}
	}()

	logsDatabaseService, err := createLogsDatabaseService(
		ctx,
		engineNamespace,
		privateGrpcPortSpec,
		logsDatabaseAttributesProvider,
		logsDatabasePodLabels,
		kubernetesManager,
	)
	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred creating the engine service")
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err = kubernetesManager.RemoveService(ctx, logsDatabaseService); err != nil {
				logrus.Errorf("Creating the logs database didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", logsDatabaseService.Name, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", logsDatabaseService.Name)
			}
		}
	}()

	shouldRemovePod = false
	shouldRemoveService = false
	return logsDatabaseService.Name, grpcPortNum, nil
}

func createLogsDatabasePod(
	ctx context.Context,
	namespace string,
	engineServiceAccount string,
	envVars map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	logsDatabaseObjAttrsProvider object_attributes_provider.KubernetesLogsDatabaseObjectAttributesProvider,
) (*apiv1.Pod, map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	// Get Pod Attributes
	logsDatabasePodAttributes, err := logsDatabaseObjAttrsProvider.ForLogsDatabasePod()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to generate pod attributes for log database")
	}
	logsDatabasePodName := logsDatabasePodAttributes.GetName().GetString()
	logsDatabasePodLabels := logsDatabasePodAttributes.GetLabels()
	logsDatabasePodLabelStrs := shared_helpers.GetStringMapFromLabelMap(logsDatabasePodLabels)
	logsDatabasePodAnnotationStrs := shared_helpers.GetStringMapFromAnnotationMap(logsDatabasePodAttributes.GetAnnotations())

	containerPorts, err := shared_helpers.GetKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the engine container ports from the private port specs")
	}

	var logsDatabaseContainerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:      varName,
			Value:     varValue,
			ValueFrom: nil,
		}
		logsDatabaseContainerEnvVars = append(logsDatabaseContainerEnvVars, envVar)
	}

	logsDatabaseContainers := []apiv1.Container{
		{
			Name:  kurtosisLogsDatabaseContainerName,
			Image: containerImage,
			Env:   logsDatabaseContainerEnvVars,
			Ports: containerPorts,
		},
	}

	var logsDatabaseVolumes []apiv1.Volume
	var logsDatabaseInitContainers []apiv1.Container

	// Create pods with engine containers and volumes in kubernetes
	pod, err := kubernetesManager.CreatePod(
		ctx,
		namespace,
		logsDatabasePodName,
		logsDatabasePodLabelStrs,
		logsDatabasePodAnnotationStrs,
		logsDatabaseInitContainers,
		logsDatabaseContainers,
		logsDatabaseVolumes,
		engineServiceAccount,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the pod with name '%s' in namespace '%s' with image '%s'", logsDatabasePodName, namespace, containerImage)
	}
	return pod, logsDatabasePodLabels, nil
}

func createLogsDatabaseService(
	ctx context.Context,
	namespace string,
	privateGrpcPortSpec *port_spec.PortSpec,
	logsDatabaseAttributesProvider object_attributes_provider.KubernetesLogsDatabaseObjectAttributesProvider,
	podMatchLabels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.Service, error) {
	logsDatabaseServiceAttributes, err := logsDatabaseAttributesProvider.ForLogsDatabaseService()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the logs database service attributes",
		)
	}
	logsDatabaseServiceName := logsDatabaseServiceAttributes.GetName().GetString()
	logsDatabaseServiceLabels := shared_helpers.GetStringMapFromLabelMap(logsDatabaseServiceAttributes.GetLabels())
	logsDatabaseServiceAnnotations := shared_helpers.GetStringMapFromAnnotationMap(logsDatabaseServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the engine pod
	servicePorts, err := shared_helpers.GetKubernetesServicePortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		consts.KurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine service's ports using the engine private port specs")
	}

	podMatchLabelStrs := shared_helpers.GetStringMapFromLabelMap(podMatchLabels)

	// Create Service
	service, err := kubernetesManager.CreateService(
		ctx,
		namespace,
		logsDatabaseServiceName,
		logsDatabaseServiceLabels,
		logsDatabaseServiceAnnotations,
		podMatchLabelStrs,
		apiv1.ServiceTypeClusterIP,
		servicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v'",
			logsDatabaseServiceName,
			namespace,
			privateGrpcPortSpec.GetNumber(),
		)
	}
	return service, nil
}

package vector

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

const (
	validatorContainerName    = "logs-aggregator-validator"
	validationCmdRetries      = 0
	validatorJobTTLSeconds    = 5
	validatorJobPollInterval  = 600 * time.Millisecond
	validatorJobPollTimeout   = 30 * time.Second
	validationSuccessExitCode = 0
	validationFailedExitCode  = 78
)

type vectorConfigurationCreator struct {
	config *VectorConfig
}

func newVectorConfigurationCreator(config *VectorConfig) *vectorConfigurationCreator {
	return &vectorConfigurationCreator{config: config}
}

func (vector *vectorConfigurationCreator) CreateConfiguration(
	ctx context.Context,
	namespaceName string,
	logsAggregatorAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*apiv1.ConfigMap, func(), error) {
	configMap, err := vector.createLogsAggregatorConfigMap(ctx, namespaceName, logsAggregatorAttrProvider, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to create config map for vector logs aggregator.")
	}

	removeConfigMapFunc := func() {
		removeCtx := context.Background()
		if err := kubernetesManager.RemoveConfigMap(removeCtx, namespaceName, configMap); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator deployment with name '%v' didn't complete successfully so we "+
					"tried to remove the config map we created, but doing so exited with an error:\n%v",
				configMap.Name,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator config map with Kubernetes name '%v' in namespace '%v'!!!!!!", configMap.Name, configMap.Namespace)
		}
	}
	shouldRemoveConfigMap := false
	defer func() {
		if shouldRemoveConfigMap {
			removeConfigMapFunc()
		}
	}()

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the validator container name")
	}

	validatorJobName := fmt.Sprintf("%s-%s", validatorContainerName, uuid)

	containers := []apiv1.Container{
		{
			Name:    validatorContainerName,
			Image:   vectorImage,
			Command: nil,
			Args:    []string{"validate", vectorConfigFilePath},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:             vectorConfigVolumeName,
					ReadOnly:         true,
					MountPath:        vectorConfigMountPath,
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
			WorkingDir: "",
			Ports:      nil,
			EnvFrom:    nil,
			Env:        nil,
			Resources: apiv1.ResourceRequirements{
				Limits:   nil,
				Requests: nil,
				Claims:   nil,
			},
			ResizePolicy:             nil,
			VolumeDevices:            nil,
			LivenessProbe:            nil,
			ReadinessProbe:           nil,
			StartupProbe:             nil,
			Lifecycle:                nil,
			TerminationMessagePath:   "",
			TerminationMessagePolicy: "",
			ImagePullPolicy:          "",
			SecurityContext:          nil,
			Stdin:                    false,
			StdinOnce:                false,
			TTY:                      false,
		},
	}

	volumes := []apiv1.Volume{
		{
			Name:         vectorConfigVolumeName,
			VolumeSource: kubernetesManager.GetVolumeSourceForConfigMap(configMap.Name),
		},
	}

	var emptyJobLabels map[string]string
	var emptyJobAnnotations map[string]string

	job, err := kubernetesManager.CreateJob(
		ctx,
		namespaceName,
		validatorJobName,
		emptyJobLabels,
		emptyJobAnnotations,
		containers,
		volumes,
		validationCmdRetries,
		validatorJobTTLSeconds,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating a job to validate logs aggregator configuration")
	}

	err = kubernetesManager.WaitForJobCompletion(ctx, job, validatorJobPollInterval, validatorJobPollTimeout)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the logs aggregator validation job to finish, even after waiting %v", validatorJobPollTimeout)
	}

	pods, err := kubernetesManager.GetPodsManagedByJob(ctx, job)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting pods for the logs aggregator validation job")
	}

	if len(pods) != 1 {
		return nil, nil, stacktrace.NewError("The logs aggregator validation job was supposed to have exactly 1 pod, this is likely a bug!")
	}

	if len(pods[0].Status.ContainerStatuses) != 1 {
		return nil, nil, stacktrace.NewError("The logs aggregator validation job was supposed to have exactly 1 container status, this is likely a bug!")
	}

	if pods[0].Status.ContainerStatuses[0].State.Terminated == nil {
		return nil, nil, stacktrace.NewError("The logs aggregator validation job was supposed to have exactly 1 terminated container, this is likely a bug!")
	}

	exitCode := pods[0].Status.ContainerStatuses[0].State.Terminated.ExitCode
	if exitCode == validationSuccessExitCode {
		shouldRemoveConfigMap = false
		return configMap, removeConfigMapFunc, nil
	}

	podName := pods[0].Name

	containerLogStream, err := kubernetesManager.GetContainerLogs(ctx, namespaceName, podName, validatorContainerName, false, false)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred getting container logs for pod %s in namespace %s for the logs aggregator validation job",
			podName,
			job.Name,
		)
	}
	defer containerLogStream.Close()

	outputBytes, err := io.ReadAll(containerLogStream)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"Failed to read container logs for pod %s in namespace %s for the logs aggregator validation job",
			podName,
			job.Name,
		)
	}

	// Vector returns a specific exit code if the validation of configurations failed
	// https://vector.dev/docs/administration/validating/#how-validation-works
	if exitCode == validationFailedExitCode {
		return nil, nil, stacktrace.NewError("The configuration provided to the logs aggregator component was invalid; errors are below:\n%s", string(outputBytes))
	}

	return nil, nil, stacktrace.NewError("Logs aggregator validation exited with non-zero status code; errors are below:\n%s", string(outputBytes))
}

func (vector *vectorConfigurationCreator) getConfigFileContent() (string, error) {
	yamlBytes, err := yaml.Marshal(vector.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error marshalling config into YAML.")
	}

	return string(yamlBytes), nil
}

func (vector *vectorConfigurationCreator) createLogsAggregatorConfigMap(
	ctx context.Context,
	namespace string,
	objAttrProvider object_attributes_provider.KubernetesLogsAggregatorObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*apiv1.ConfigMap,
	error,
) {
	configMapAttrProvider, err := objAttrProvider.ForLogsAggregatorConfigMap()
	if err != nil {
		return nil, err
	}
	name := configMapAttrProvider.GetName().GetString()
	labels := shared_helpers.GetStringMapFromLabelMap(configMapAttrProvider.GetLabels())
	annotations := shared_helpers.GetStringMapFromAnnotationMap(configMapAttrProvider.GetAnnotations())

	vectorConfigStr, err := vector.getConfigFileContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating vector config.")
	}

	configMap, err := kubernetesManager.CreateConfigMap(
		ctx,
		namespace,
		name,
		labels,
		annotations,
		map[string]string{
			vectorConfigFileName: vectorConfigStr,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating config map for vector log aggregator config.")
	}

	return configMap, nil
}

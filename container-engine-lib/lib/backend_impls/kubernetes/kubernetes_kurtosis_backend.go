package kubernetes

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/sirupsen/logrus"
)

const (
	kurtosisEngineNamespace    = "kurtosis-namespace"
	numKurtosisEngineReplicas  = 1
	storageClass               = "standard"
	defaultQuantity            = "10Gi"
	defaultHostPathInMinikube  = "/host/data/engine-data"
	externalServiceType        = "LoadBalancer"
	numReplicasToStopContainer = 0

	// Engine container port number string parsing constants
	hostMachinePortNumStrParsingBase = 10
	hostMachinePortNumStrParsingBits = 16

	shouldCleanRunningEngineContainers = false
)

type KubernetesKurtosisBackend struct {
	kubernetesManager *kubernetes_manager.KubernetesManager
}

func NewKubernetesKurtosisBackend(log *logrus.Logger, k8sManager *kubernetes_manager.KubernetesManager) *KubernetesKurtosisBackend {
	return &KubernetesKurtosisBackend{
		kubernetesManager: k8sManager,
	}
}

/*
var engineLabels = map[string]string{
	// TODO don't use a shared place for both Docker & Kubernetes for this; each backend should have its own labels
	forever_constants.AppIDLabel:         forever_constants.AppIDValue,
	forever_constants.ContainerTypeLabel: forever_constants.ContainerType_EngineServer,
}

type KubernetesKurtosisBackend struct {
	log *logrus.Logger

	kubernetesManager *kubernetes_manager.KubernetesManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewKubernetesKurtosisBackendCore(log *logrus.Logger, k8sManager *kubernetes_manager.KubernetesManager, objAttrsProvider schema.ObjectAttributesProvider) *KubernetesKurtosisBackendCore {
	return &KubernetesKurtosisBackendCore{
		log: log,

		kubernetesManager: k8sManager,
		objAttrsProvider:  objAttrsProvider,
	}
}

func (backendCore KubernetesKurtosisBackendCore) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	imageOrgAndRepo string,
	envVars map[string]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	// getting the object attributes for the engine server
	engineAttrs, err := backendCore.objAttrsProvider.ForEngineServer(listenPortNum) // TODO we should probably create a new function for labels that make sense for kubernetes deployment
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	// getting the object attributes for the engine server
	engineAttrsForPod, err := backendCore.objAttrsProvider.ForEngineServer(listenPortNum) // TODO we should probably create a new function for labels that make sense for kubernetes pod
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	engineDataDirpathOnHostMachine = defaultHostPathInMinikube

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	// checking if the kurtosis namespace already exists and creating it otherwise
	kurtosisNamespaceList, err := backendCore.kubernetesManager.GetNamespacesByLabels(ctx, engineLabels)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to get the kurtosis engine namespace by labels '%+v'",
			engineAttrs.GetLabels())
	}
	if len(kurtosisNamespaceList.Items) == 0 {
		_, err = backendCore.kubernetesManager.CreateNamespace(ctx, kurtosisEngineNamespace, engineLabels)
		if err != nil {
			return nil, 0, stacktrace.Propagate(
				err,
				"An error occurred when trying to create the kurtosis engine namespace to be named '%v'",
				kurtosisEngineNamespace)
		}
	}

	// creating persistent volume
	_, err = backendCore.kubernetesManager.CreatePersistentVolume(ctx, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, engineDataDirpathOnHostMachine, storageClass)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to create the persistent volume to be named '%v'",
			engineAttrs.GetName())
	}

	// creating persistent volume claim
	_, err = backendCore.kubernetesManager.CreatePersistentVolumeClaim(ctx, kurtosisEngineNamespace, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, storageClass)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to create the persistent volume claim to be named '%v'",
			engineAttrs.GetName())
	}

	// defining the volumes for the deployment
	volumes := []apiv1.Volume{
		{
			Name: engineAttrs.GetName(),
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: engineAttrs.GetName(),
				},
			},
		},
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      engineAttrs.GetName(),
			MountPath: kurtosis_backend.EngineDataDirpathOnEngineServerContainer,
		},
	}

	// creating deployment
	_, err = backendCore.kubernetesManager.CreateDeployment(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), engineAttrsForPod.GetLabels(), containerImageAndTag, numKurtosisEngineReplicas, volumes, volumeMounts, envVars, engineAttrs.GetName())
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred while creating the deployment with name '%s' in namespace '%s' with image '%s'", engineAttrs.GetName(), kurtosisEngineNamespace, containerImageAndTag)
	}

	// creating service
	service, err := backendCore.kubernetesManager.CreateService(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), externalServiceType, int32(listenPortNum), int32(listenPortNum))
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s'", engineAttrs.GetName(), kurtosisEngineNamespace)
	}

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, hostMachinePortNumStrParsingBase, hostMachinePortNumStrParsingBits)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			hostMachinePortNumStrParsingBase,
			hostMachinePortNumStrParsingBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return publicIpAddr, publicPortNumUint16, nil
}

// TODO Replace with a GetEngine command to get information about the engine
func (backendCore KubernetesKurtosisBackendCore) GetEnginePublicIPAndPort(
	ctx context.Context,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultIsEngineStopped bool,
	resultErr error,
) {
	deploymentList, err := backendCore.kubernetesManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return nil, 0, false, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine deployment with labels '%+v' in namespace '%s'", engineLabels, kurtosisEngineNamespace)
	}

	var deployments []v1.Deployment
	for _, deployment := range deploymentList.Items {
		if *deployment.Spec.Replicas > 0 {
			deployments = append(deployments, deployment)
		}
	}

	numRunningEngines := len(deployments)
	if numRunningEngines > 1 {
		return nil, 0, false, stacktrace.NewError("Cannot report engine status because we found '%v' running Kurtosis engines; this is very strange as there should never be more than one", numRunningEngines)
	}

	if numRunningEngines == 0 {
		return nil, 0, true, nil
	}

	engineDeployment := deployments[0]

	service, err := backendCore.kubernetesManager.GetServiceByName(ctx, kurtosisEngineNamespace, engineDeployment.Name)
	if err != nil {
		return nil, 0, false, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine service with name '%s' in namespace '%s'", engineDeployment.Name, kurtosisEngineNamespace)
	}

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, hostMachinePortNumStrParsingBase, hostMachinePortNumStrParsingBits)
	if err != nil {
		return nil, 0, false, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			hostMachinePortNumStrParsingBase,
			hostMachinePortNumStrParsingBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return publicIpAddr, publicPortNumUint16, false, nil
}

func (backendCore KubernetesKurtosisBackendCore) StopEngine(ctx context.Context) error {
	err := backendCore.kubernetesManager.UpdateDeploymentReplicas(ctx, kurtosisEngineNamespace, engineLabels, int32(numReplicasToStopContainer))
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while trying to stop the engine server with labels '%+v'", engineLabels)
	}

	return nil
}

func (backendCore KubernetesKurtosisBackendCore) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	deploymentsList, err := backendCore.kubernetesManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to get the deployment with labels '%+v'", engineLabels)
	}

	successfullyDestroyedEngineNames := []string{}
	removeEngineErrors := []error{}

	if len(deploymentsList.Items) > 0 {
		for _, deployment := range deploymentsList.Items {
			if !backendCore.checkIfContainerIsRunning(deployment) {
				err = backendCore.cleanEngineServer(ctx, deployment.Name)
				if err != nil {
					removeEngineErrors = append(removeEngineErrors, err)
				} else {
					successfullyDestroyedEngineNames = append(successfullyDestroyedEngineNames, deployment.Name)
				}
			}
		}
	}

	return successfullyDestroyedEngineNames, removeEngineErrors, nil
}

func (backendCore KubernetesKurtosisBackendCore) checkIfContainerIsRunning(deployment v1.Deployment) bool {
	return *deployment.Spec.Replicas > 0
}

func (backendCore KubernetesKurtosisBackendCore) cleanEngineServer(ctx context.Context, name string) error {
	err := backendCore.kubernetesManager.RemoveDeployment(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the deployment from the engine with name '%s'", name)
	}

	err = backendCore.kubernetesManager.RemovePersistentVolumeClaim(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the persistent volume claim from the engine with name '%s'", name)
	}

	err = backendCore.kubernetesManager.RemovePersistentVolume(ctx, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the persistent volume from the engine with name '%s'", name)
	}

	err = backendCore.kubernetesManager.RemoveService(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the service from the engine with name '%s'", name)
	}

	return nil
}

 */

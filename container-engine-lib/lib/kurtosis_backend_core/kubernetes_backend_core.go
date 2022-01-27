package kurtosis_backend_core

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/k8s_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend_core/args"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"strconv"
)

const (
	kurtosisEngineNamespace   = "kurtosis-namespace"
	kurtosisEngineReplicas    = 1
	storageClass              = "standard"
	defaultQuantity           = "10Gi"
	defaultHostPathInMinikube = "/host/data/engine-data"
	externalServiceType       = "LoadBalancer"
	zeroReplicas              = 0
)

type KurtosisKubernetesBackendCore struct {
	log *logrus.Logger

	k8sManager *k8s_manager.K8sManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewKurtosisKubernetesBackendCore(log *logrus.Logger, k8sManager *k8s_manager.K8sManager, objAttrsProvider schema.ObjectAttributesProvider) *KurtosisKubernetesBackendCore {
	return &KurtosisKubernetesBackendCore{
		log:              log,
		k8sManager:       k8sManager,
		objAttrsProvider: objAttrsProvider,
	}
}

func (kkb KurtosisKubernetesBackendCore) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	containerImage string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	// getting the object attributes for the engine server
	engineAttrs, err := kkb.objAttrsProvider.ForEngineServer(listenPortNum)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	engineDataDirpathOnHostMachine = defaultHostPathInMinikube

	// getting the args object for the engine server container
	argsObj, err := args.NewEngineServerArgs(
		listenPortNum,
		logLevel.String(),
		imageVersionTag,
		engineDataDirpathOnHostMachine,
	)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred creating the engine server args")
	}

	// getting the envVars for the engine server container
	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	// checking if the kurtosis namespace already exists and creating it otherwise
	kurtosisNamespaceList, err := kkb.k8sManager.GetNamespacesByLabels(ctx,engineLabels)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to get the kurtosis engine namespace by labels '%v'",
			engineAttrs.GetLabels())
	}
	if len(kurtosisNamespaceList.Items) == 0 {
		_, err = kkb.k8sManager.CreateNamespace(ctx, kurtosisEngineNamespace, engineLabels)
		if err != nil {
			return nil, 0, stacktrace.Propagate(
				err,
				"An error occurred when trying to create the kurtosis engine namespace to be named '%v'",
				kurtosisEngineNamespace)
		}
	}

	// creating persistent volume
	_, err = kkb.k8sManager.CreatePersistentVolume(ctx, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, argsObj.EngineDataDirpathOnHostMachine, storageClass)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to create the persistent volume to be named '%v'",
			engineAttrs.GetName())
	}

	// creating persistent volume claim
	_, err = kkb.k8sManager.CreatePersistentVolumeClaim(ctx, kurtosisEngineNamespace, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, storageClass)
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
			MountPath: EngineDataDirpathOnEngineServerContainer,
		},
	}

	// creating deployment
	_, err = kkb.k8sManager.CreateDeployment(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), containerImageAndTag, kurtosisEngineReplicas, volumes, volumeMounts, envVars, engineAttrs.GetName())
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	// creating service
	service, err := kkb.k8sManager.CreateService(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), externalServiceType, int32(listenPortNum), int32(listenPortNum))
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumParsingBase, publicPortNumParsingUintBits)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumParsingBase,
			publicPortNumParsingUintBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return publicIpAddr, publicPortNumUint16, nil
}

func (kkb KurtosisKubernetesBackendCore) StopEngine(ctx context.Context) error {
	err := kkb.k8sManager.UpdateDeploymentReplicas(ctx, kurtosisEngineNamespace, engineLabels, int32(zeroReplicas))
	if err != nil {
		stacktrace.Propagate(err, "an error occurred while trying to stop the engine server with labels %v", engineLabels)
	}

	return nil
}

func (kkb KurtosisKubernetesBackendCore) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	deploymentsList, err := kkb.k8sManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "an error occurred while trying to clean the the stopped engine containers with labels %v", engineLabels)
	}

	successfullyDestroyedEngineNames := []string{}
	removeEngineErrors := []error{}

	if len(deploymentsList.Items) > 0 {
		for _, deployment := range deploymentsList.Items {
			if !kkb.checkIfContainerisRunning(deployment) {
				err = kkb.cleanEngineServer(ctx, deployment.Name)
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

func (kkb KurtosisKubernetesBackendCore) checkIfContainerisRunning(deployment v1.Deployment) bool {
	return *deployment.Spec.Replicas > 0
}

func (kkb KurtosisKubernetesBackendCore) cleanEngineServer(ctx context.Context, name string) error {
	err := kkb.k8sManager.RemoveDeployment(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while trying to delete the deployment from the engine with name %s", name)
	}

	err = kkb.k8sManager.RemovePersistentVolumeClaim(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while trying to delete the persistent volume claim from the engine with name %s", name)
	}

	err = kkb.k8sManager.RemovePersistentVolume(ctx, name)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while trying to delete the persistent volume from the engine with name %s", name)
	}

	err = kkb.k8sManager.RemoveService(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while trying to delete the service from the engine with name %s", name)
	}

	return nil
}

func (kkb KurtosisKubernetesBackendCore) GetEngineStatus(
	ctx context.Context,
) (engineStatus string, ipAddr net.IP, portNum uint16, err error) {
	deploymentList, err := kkb.k8sManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine deployments")
	}

	var deployments []v1.Deployment
	for _, deployment := range deploymentList.Items {
		if *deployment.Spec.Replicas > 0 {
			deployments = append(deployments, deployment)
		}
	}

	numRunningEngines := len(deployments)
	if numRunningEngines > 1 {
		return "", nil, 0, stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engines; this is very strange as there should never be more than one", numRunningEngines)
	}

	if numRunningEngines == 0 {
		return EngineStatus_Stopped, nil, 0, nil
	}

	engineDeployment := deployments[0]

	service, err:=kkb.k8sManager.GetServiceByName(ctx, kurtosisEngineNamespace, engineDeployment.Name)

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumParsingBase, publicPortNumParsingUintBits)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumParsingBase,
			publicPortNumParsingUintBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return "", publicIpAddr, publicPortNumUint16, nil
}

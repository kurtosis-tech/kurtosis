package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

type kubernetesVolumeWithClaim struct {
	VolumeClaimName string
}

func (volumeAndClaim *kubernetesVolumeWithClaim) GetVolume() *apiv1.Volume {
	return &apiv1.Volume{
		Name: volumeAndClaim.VolumeClaimName,
		VolumeSource: apiv1.VolumeSource{
			HostPath:             nil,
			EmptyDir:             nil,
			GCEPersistentDisk:    nil,
			AWSElasticBlockStore: nil,
			GitRepo:              nil,
			Secret:               nil,
			NFS:                  nil,
			ISCSI:                nil,
			Glusterfs:            nil,
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeAndClaim.VolumeClaimName,
				ReadOnly:  false,
			},
			RBD:                  nil,
			FlexVolume:           nil,
			Cinder:               nil,
			CephFS:               nil,
			Flocker:              nil,
			DownwardAPI:          nil,
			FC:                   nil,
			AzureFile:            nil,
			ConfigMap:            nil,
			VsphereVolume:        nil,
			Quobyte:              nil,
			AzureDisk:            nil,
			PhotonPersistentDisk: nil,
			Projected:            nil,
			PortworxVolume:       nil,
			ScaleIO:              nil,
			StorageOS:            nil,
			CSI:                  nil,
			Ephemeral:            nil,
		},
	}
}

func (volumeAndClaim *kubernetesVolumeWithClaim) GetVolumeMount(mountPath string) *apiv1.VolumeMount {
	return &apiv1.VolumeMount{
		Name:             volumeAndClaim.VolumeClaimName,
		ReadOnly:         false,
		MountPath:        mountPath,
		SubPath:          "",
		MountPropagation: nil,
		SubPathExpr:      "",
	}
}

func preparePersistentDirectoriesResources(
	ctx context.Context,
	namespace string,
	objAttributeProviders object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
	serviceMountpointsToPersistentKey map[string]service_directory.PersistentDirectory,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (map[string]*kubernetesVolumeWithClaim, error) {
	shouldDeleteVolumeClaims := true
	volumeClaimsCreated := map[string]*apiv1.PersistentVolumeClaim{}

	persistentVolumesAndClaims := map[string]*kubernetesVolumeWithClaim{}

	for dirPath, persistentDirectory := range serviceMountpointsToPersistentKey {
		volumeAttrs, err := objAttributeProviders.ForSinglePersistentDirectoryVolume(persistentDirectory.PersistentKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the labels for persist service directory '%s'", persistentDirectory.PersistentKey)
		}

		volumeName := volumeAttrs.GetName().GetString()
		volumeLabelsStrs := map[string]string{}
		for key, value := range volumeAttrs.GetLabels() {
			volumeLabelsStrs[key.GetString()] = value.GetString()
		}

		persistentVolumeSize := int64(persistentDirectory.Size)

		// This claim works with a dynamic driver - it will spin up its own volume - the volume will get deleted when said claims is deleted
		var persistentVolumeClaim *apiv1.PersistentVolumeClaim
		if persistentVolumeClaim, err = kubernetesManager.GetPersistentVolumeClaim(ctx, namespace, volumeName); err != nil {
			persistentVolumeClaim, err = kubernetesManager.CreatePersistentVolumeClaim(ctx, namespace, volumeName, volumeLabelsStrs, persistentVolumeSize)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating the persistent volume claim for '%s'", persistentDirectory.PersistentKey)
			}
			volumeClaimsCreated[persistentVolumeClaim.Name] = persistentVolumeClaim
		}

		persistentVolumesAndClaims[dirPath] = &kubernetesVolumeWithClaim{
			VolumeClaimName: persistentVolumeClaim.Name,
		}
	}

	defer func() {
		if !shouldDeleteVolumeClaims {
			return
		}
		for volumeClaimNameStr := range volumeClaimsCreated {
			// Background context so we still run this even if the input context was cancelled
			if err := kubernetesManager.RemovePersistentVolumeClaim(context.Background(), namespace, volumeClaimNameStr); err != nil {
				logrus.Warnf(
					"Creating persistent directory volumes didn't complete successfully so we tried to delete volume claim '%v' that we created, but doing so threw an error:\n%v",
					volumeClaimNameStr,
					err,
				)
				logrus.Warnf("You'll need to clean up volume claim '%v' manually!", volumeClaimNameStr)
			}
		}
	}()

	shouldDeleteVolumeClaims = false
	return persistentVolumesAndClaims, nil
}

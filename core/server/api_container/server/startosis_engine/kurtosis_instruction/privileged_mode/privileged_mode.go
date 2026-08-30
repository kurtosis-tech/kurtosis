package privileged_mode

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
)

func ValidateServiceConfig(serviceConfig *service.ServiceConfig, allowPrivilegedMode bool, backendType args.KurtosisBackendType) *startosis_errors.InterpretationError {
	if serviceConfig == nil || (!serviceConfig.GetPrivileged() && len(serviceConfig.GetBindMounts()) == 0 && !serviceConfig.GetHostPIDNamespace() && !serviceConfig.GetHostCgroupNamespace()) {
		return nil
	}
	if backendType != args.KurtosisBackendType_Docker {
		return startosis_errors.NewInterpretationError(
			"ServiceConfig requested privileged=true, bind_mounts, host_pid_namespace=true, or host_cgroup_namespace=true, but these settings are Docker-only and are not supported on the %s backend",
			backendType.String(),
		)
	}
	if !allowPrivilegedMode {
		return startosis_errors.NewInterpretationError(
			"ServiceConfig requested privileged=true, bind_mounts, host_pid_namespace=true, or host_cgroup_namespace=true, but this run did not opt in. Pass --privileged on the CLI, or set allow-privileged-mode: true in kurtosis-config.yml",
		)
	}
	return nil
}

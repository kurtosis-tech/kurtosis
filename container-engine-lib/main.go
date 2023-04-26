package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	if err := runMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

// You can comment out various sections to test various parts of the lib
func runMain() error {
	/*
		if err := runDockerManagerTesting(); err != nil {
			return err
		}

	*/

	/*
		if err := runKubernetesManagerTesting(); err != nil {
			return err
		}
	*/

	if err := runKurtosisBackendTesting(); err != nil {
		return err
	}

	return nil
}

/*
func runDockerManagerTesting() error {
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}
	dockerManager := docker_manager.NewDockerManager(dockerClient)

	// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv Arbitrary logic goes here vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	result, err := dockerManager.GetContainersByLabels(ctx, map[string]string{}, false)
	if err != nil {
		return err
	}
	fmt.Println(result)

	return nil
}
*/

/*
func runKubernetesManagerTesting() error {
	ctx := context.Background()
	kubeConfigFileFilepath := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFileFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating kubernetes configuration from flags in file '%v'", kubeConfigFileFilepath)
	}
	clientSet, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to create kubernetes client set using Kubernetes config '%+v', instead a non nil error was returned", kubernetesConfig)
	}
	kubernetesManager := kubernetes_manager.NewKubernetesManager(clientSet, kubernetesConfig)

	// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv Arbitrary logic goes here vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	_, err = kubernetesManager.GetNamespace(ctx, "TODO")
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the name space")
	}

	return nil
}
*/

// Can comment which backend you want to use
func runKurtosisBackendTesting() error {
	ctx := context.Background()

	backend, err := backend_creator.GetLocalDockerKurtosisBackend(nil)
	if err != nil {
		return err
	}

	serializedArgs := map[string]string{
		"SERIALIZED_ARGS": `{"grpcListenPortNum":9710,"grpcProxyListenPortNum":9711,"logLevelStr":"debug","imageVersionTag":"1.29.0","metricsUserId":"552f","didUserAcceptSendingMetrics":false,"kurtosisBackendType":"docker","kurtosisBackendConfig":{}}`,
	}

	engine, err := backend.CreateEngine(
		ctx,
		"kurtosistech/engine",
		"0.49.5",
		9710,
		9711,
		serializedArgs,
	)
	if err != nil {
		return err
	}
	logrus.Infof("Engine 1 info: %+v", engine)

	/*engineFil := &engine_object.EngineFilters{
		UUIDs: map[engine_object.EngineGUID]bool{
			engine.GetUUID(): true,
		},
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Running: true,
		},
	}
	stoppedEngineGuids, erroredEngineGuids, err := backend.StopEngines(ctx, engineFil)
	if err != nil {
		return err
	}
	logrus.Infof("Successfull stopped engines: %+v", stoppedEngineGuids)
	logrus.Infof("Errored stopped engines: %+v", erroredEngineGuids)
	*/

	serializedArgs2 := map[string]string{
		"SERIALIZED_ARGS": `{"grpcListenPortNum":9810,"grpcProxyListenPortNum":9811,"logLevelStr":"debug","imageVersionTag":"1.29.0","metricsUserId":"552f","didUserAcceptSendingMetrics":false,"kurtosisBackendType":"docker","kurtosisBackendConfig":{}}`,
	}

	engine2, err := backend.CreateEngine(
		ctx,
		"kurtosistech/kurtosis-engine-server",
		"1.29.0",
		9810,
		9811,
		serializedArgs2,
	)
	if err != nil {
		return err
	}
	logrus.Infof("Engine 2 info: %+v", engine2)

	/*
		engineFil2 := &engine_object.EngineFilters{
			UUIDs: map[engine_object.EngineGUID]bool{
				engine.GetUUID(): true,
				engine2.GetUUID(): true,
			},
		}
		destroyedEngineGuids, erroredDestroyedEngineGuids, err := backend.DestroyEngines(ctx, engineFil2)
		if err != nil {
			return err
		}
		logrus.Infof("Successfull destroyed engines: %+v", destroyedEngineGuids)
		logrus.Infof("Errored destroyed engines: %+v", erroredDestroyedEngineGuids)

		/*
			enclaveID := enclave2.EnclaveID("enclave-for-test")
			enclave, err := backend.CreateEnclave(
				ctx,
				enclaveID,
				false,
			)
			if err != nil {
				return err
			}
			logrus.Infof("Enclave info: %+v", enclave)

			userServiceName := service.ServiceName("user-service-test")
			serviceNames := map[service.ServiceName]bool {
				userServiceName: true,
			}
			successfulUserServiceRegistrations, erroredUserServiceRegistrations, err := backend.RegisterUserServices(
				ctx,
				enclaveID,
				serviceNames,
			)
			if err != nil {
				return err
			}
			logrus.Infof("Successfull user service registrations: %+v", successfulUserServiceRegistrations)
			logrus.Infof("Errored user service registrations: %+v", erroredUserServiceRegistrations)

			serviceConfig := service.NewServiceConfig(
				"alpine:3.17",
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				0,
				0,
			)

			serviceToStart := map[service.ServiceUUID]*service.ServiceConfig{
				successfulUserServiceRegistrations[userServiceName].GetUUID(): serviceConfig,
			}

			successfulUserServiceStarted, erroredUserServiceStarted, err := backend.StartUserServices(
				ctx,
				enclaveID,
				serviceToStart,
			)
			if err != nil {
				return err
			}
			logrus.Infof("Successfull user service started: %+v", successfulUserServiceStarted)
			logrus.Infof("Errored user service started: %+v", erroredUserServiceStarted)
	*/

	/*
		_, err := lib.GetCLIKubernetesKurtosisBackend(ctx)
		if err != nil {
			return err
		}
	*/

	// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv Arbitrary logic goes here vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	// enclaveId := enclave.EnclaveID("test")  // TODO Make this whatever you need
	// serviceUuid := service.ServiceUUID("TODO")
	/*
		results, err := backend.CreateFilesArtifactExpansion(ctx, "test", "TODO", "/foo/bar")
		if err != nil {
			return err
		}
		fmt.Println(results)

	*/

	return nil
}

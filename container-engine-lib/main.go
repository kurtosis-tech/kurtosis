package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
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
	kubernetesManager.GetNamespace(ctx, "TODO")

	return nil
}

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
		"kurtosistech/kurtosis-engine-server",
		"1.29.0",
		9710,
		9810,
		serializedArgs,
	)
	if err != nil {
		return err
	}

	logrus.Infof("Engine info: %+v", engine)

	/*
		_, err := lib.GetCLIKubernetesKurtosisBackend(ctx)
		if err != nil {
			return err
		}
	*/


	// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv Arbitrary logic goes here vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	// enclaveId := enclave.EnclaveID("test")  // TODO Make this whatever you need
	// serviceGuid := service.ServiceGUID("TODO")
	/*
		results, err := backend.CreateFilesArtifactExpansion(ctx, "test", "TODO", "/foo/bar")
		if err != nil {
			return err
		}
		fmt.Println(results)

	*/


	return nil
}


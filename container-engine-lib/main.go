package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
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

// You can comment out various sections to test various things
func runMain() error {
	/*
	if err := runKubernetesKurtosisBackendTesting(); err != nil {
		return err
	}
	 */

	/*
		if err := runKubernetesManagerTesting(); err != nil {
			return err
		}
	*/

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

	// TODO replace this with whatever you want to test
	kubernetesManager.GetNamespace(ctx, "TODO")

	return nil
}

func runKubernetesKurtosisBackendTesting() error {
	ctx := context.Background()
	backend, err := lib.GetCLIKubernetesKurtosisBackend(ctx)
	if err != nil {
		return err
	}

	// TODO replace this with whatever you want to test
	enclaveId := enclave.EnclaveID("TODO")  // TODO Make this whatever you need
	serviceGuid := service.ServiceGUID("TODO")
	filters := &service.ServiceFilters{
		GUIDs:    map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	results, err := backend.GetUserServices(ctx, enclaveId, filters)
	if err != nil {
		return err
	}
	fmt.Println(results)


	return nil
}

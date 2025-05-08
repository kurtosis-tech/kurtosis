package starlark_script_creator

import (
	"fmt"
	"sort"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	// eg. plan.upload_files(src = "./data/project", name = "web-volume0")
	uploadFilesLinesFmtStr = "plan.upload_files(src = \"%s\", name = \"%s\")"

	// eg. plan.add_service(name="web", config=ServiceConfig(...))
	addServiceLinesFmtStr = "plan.add_service(name = \"%s\", config = %s)"

	defRunStr = "def run(plan):\n"

	newStarlarkLineFmtStr = "    %s\n"
)

var CyclicalDependencyError = stacktrace.NewError("A cycle was detected in the service dependency graph.")

type StarlarkServiceConfig *kurtosis_type_constructor.KurtosisValueTypeDefault

// Creates a starlark script based on starlark ServiceConfigs, the service dependency graph, and files artifacts to upload
func CreateStarlarkScript(
	serviceNameToStarlarkServiceConfig map[string]StarlarkServiceConfig,
	serviceDependencyGraph map[string]map[string]bool,
	servicesToFilesArtifactsToUpload map[string]map[string]string) (string, error) {
	starlarkLines := []string{}

	// Add add_service instructions in an order that respects [serviceDependencyGraph] determined by 'depends_on' keys in Compose
	sortedServices, err := sortServicesBasedOnDependencies(serviceDependencyGraph)
	if err != nil {
		return "", err
	}
	for _, serviceName := range sortedServices {
		// upload_files artifacts for service
		// get and sort keys first for deterministic order
		filesArtifactsToUpload := servicesToFilesArtifactsToUpload[serviceName]
		sortedRelativePaths := []string{}
		for relativePath := range filesArtifactsToUpload {
			sortedRelativePaths = append(sortedRelativePaths, relativePath)
		}
		sort.Strings(sortedRelativePaths)
		for _, relativePath := range sortedRelativePaths {
			filesArtifactName := filesArtifactsToUpload[relativePath]
			uploadFilesLine := fmt.Sprintf(uploadFilesLinesFmtStr, relativePath, filesArtifactName)
			starlarkLines = append(starlarkLines, uploadFilesLine)
		}

		// add_service
		starlarkServiceConfig := *serviceNameToStarlarkServiceConfig[serviceName]
		addServiceLine := fmt.Sprintf(addServiceLinesFmtStr, serviceName, starlarkServiceConfig.String())
		starlarkLines = append(starlarkLines, addServiceLine)
	}

	script := defRunStr
	for _, line := range starlarkLines {
		script += fmt.Sprintf(newStarlarkLineFmtStr, line)
	}
	return script, nil
}

func AppendKwarg(kwargs []starlark.Tuple, argName string, argValue starlark.Value) []starlark.Tuple {
	tuple := []starlark.Value{
		starlark.String(argName),
		argValue,
	}
	return append(kwargs, tuple)
}

// Returns list of service names in an order that respects dependencies by performing a topological sort
// Returns error if cyclical dependency is detected
// o(n^2) but simpler variation of Kahns algorithm https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm
// To ensure a deterministic sort, ties are broken lexicographically based on service name
func sortServicesBasedOnDependencies(serviceDependencyGraph map[string]map[string]bool) ([]string, error) {
	initServices := []string{} // start services with services that have no dependencies
	for service, dependencies := range serviceDependencyGraph {
		if len(dependencies) == 0 {
			initServices = append(initServices, service)
		}
	}

	sortedServices := []string{}
	queue := []string{}
	sort.Strings(initServices)
	queue = append(queue, initServices...)

	for len(queue) > 0 {
		processedService := queue[0]
		queue = queue[1:]
		sortedServices = append(sortedServices, processedService)
		delete(serviceDependencyGraph, processedService)

		servicesToQueue := []string{}
		for service, dependencies := range serviceDependencyGraph {
			// Remove processedService if it was as a dependency
			if dependencies[processedService] {
				delete(dependencies, processedService)

				// add service to queue if all of its dependencies have been processed
				if len(dependencies) == 0 {
					servicesToQueue = append(servicesToQueue, service)
				}
			}
		}

		sort.Strings(servicesToQueue)
		queue = append(queue, servicesToQueue...)
	}

	// If there are still dependencies that need to be processed, a cycle exists
	if len(serviceDependencyGraph) > 0 {
		return nil, CyclicalDependencyError
	}

	return sortedServices, nil
}

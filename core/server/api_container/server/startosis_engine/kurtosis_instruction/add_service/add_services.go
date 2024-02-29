package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
	"sync"
)

const (
	AddServicesBuiltinName = "add_services"

	ConfigsArgName = "configs"
)

func NewAddServices(
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: AddServicesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ConfigsArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication with Interpret
						_, ok := value.(*starlark.Dict)
						if !ok {
							return startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", ConfigsArgName, reflect.TypeOf(value))
						}
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &AddServicesCapabilities{
				serviceNetwork:               serviceNetwork,
				runtimeValueStore:            runtimeValueStore,
				packageId:                    packageId,
				packageContentProvider:       packageContentProvider,
				packageReplaceOptions:        packageReplaceOptions,
				serviceConfigs:               nil, // populated at interpretation time
				interpretationTimeValueStore: interpretationTimeValueStore,

				resultUuids:     map[service.ServiceName]string{}, // populated at interpretation time
				readyConditions: nil,                              // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			// adding the entire config as a representative arg is kind of sad here as it might clutter the output,
			// but we don't really the choice
			ConfigsArgName: true,
		},
	}
}

type AddServicesCapabilities struct {
	serviceNetwork               service_network.ServiceNetwork
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	packageId                    string
	packageContentProvider       startosis_packages.PackageContentProvider
	packageReplaceOptions        map[string]string
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	serviceConfigs map[service.ServiceName]*service.ServiceConfig

	readyConditions map[service.ServiceName]*service_config.ReadyCondition

	resultUuids map[service.ServiceName]string
}

func (builtin *AddServicesCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	ServiceConfigsDict, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, ConfigsArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ConfigsArgName)
	}
	serviceConfigs, readyConditions, interpretationErr := validateAndConvertConfigsAndReadyConditions(
		builtin.serviceNetwork,
		ServiceConfigsDict,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		builtin.packageId,
		builtin.packageContentProvider,
		builtin.packageReplaceOptions,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.serviceConfigs = serviceConfigs
	builtin.readyConditions = readyConditions

	resultUuids, returnValue, interpretationErr := makeAndPersistAddServicesInterpretationReturnValue(builtin.serviceConfigs, builtin.runtimeValueStore, builtin.interpretationTimeValueStore)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.resultUuids = resultUuids
	return returnValue, nil
}

func (builtin *AddServicesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	for serviceName, serviceConfig := range builtin.serviceConfigs {
		if err := validateSingleService(validatorEnvironment, serviceName, serviceConfig); err != nil {
			return err
		}
	}
	return nil
}

func (builtin *AddServicesCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	renderedServiceConfigs := make(map[service.ServiceName]*service.ServiceConfig, len(builtin.serviceConfigs))
	parallelism, ok := ctx.Value(startosis_constants.ParallelismParam).(int)
	if !ok {
		return "", stacktrace.NewError("An error occurred when getting parallelism level from execution context")
	}
	for serviceName, serviceConfig := range builtin.serviceConfigs {
		renderedServiceName, renderedServiceConfig, err := replaceMagicStrings(builtin.runtimeValueStore, serviceName, serviceConfig)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred replacing a magic string in '%s' instruction arguments for service: '%s'. Execution cannot proceed", AddServicesBuiltinName, serviceName)
		}
		renderedServiceConfigs[renderedServiceName] = renderedServiceConfig
	}

	serviceToUpdate := map[service.ServiceName]*service.ServiceConfig{}
	serviceToCreate := map[service.ServiceName]*service.ServiceConfig{}
	for serviceName, serviceConfig := range renderedServiceConfigs {
		exist, err := builtin.serviceNetwork.ExistServiceRegistration(serviceName)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting service registration for service '%s'", serviceName)
		}
		if exist {
			serviceToUpdate[serviceName] = serviceConfig
		} else {
			serviceToCreate[serviceName] = serviceConfig
		}
	}

	updatedServices, failedToBeUpdatedServices, err := builtin.serviceNetwork.UpdateServices(ctx, serviceToUpdate, parallelism)
	if err != nil {
		var allServiceNames []string
		for serviceName := range serviceToUpdate {
			allServiceNames = append(allServiceNames, string(serviceName))
		}
		return "", stacktrace.Propagate(err, "Unexpected error occurred updating the following batch of services: %s", strings.Join(allServiceNames, ", "))
	}

	startedServices, failedToBeStartedServices, err := builtin.serviceNetwork.AddServices(ctx, serviceToCreate, parallelism)
	if err != nil {
		var allServiceNames []string
		for serviceName := range serviceToCreate {
			allServiceNames = append(allServiceNames, string(serviceName))
		}
		return "", stacktrace.Propagate(err, "Unexpected error occurred starting the following batch of services: %s", strings.Join(allServiceNames, ", "))
	}
	if len(failedToBeStartedServices) > 0 || len(failedToBeUpdatedServices) > 0 {
		var failedServiceNames []service.ServiceName
		for failedServiceName := range failedToBeStartedServices {
			failedServiceNames = append(failedServiceNames, failedServiceName)
		}
		for failedServiceName := range failedToBeUpdatedServices {
			failedServiceNames = append(failedServiceNames, failedServiceName)
		}
		return "", stacktrace.NewError("Some errors occurred starting or updating the following services: '%v'. The entire batch was rolled back an no service was started. Errors were:\nService creations: %v\nService Updates: %v", failedServiceNames, failedToBeStartedServices, failedToBeUpdatedServices)
	}
	startedAndUpdatedService := map[service.ServiceName]*service.Service{}
	for startedServiceName, startedService := range startedServices {
		startedAndUpdatedService[startedServiceName] = startedService
	}
	for updatedServiceName, updatedService := range updatedServices {
		startedAndUpdatedService[updatedServiceName] = updatedService
	}
	shouldDeleteAllStartedServices := true

	//TODO we should move the readiness check functionality to the default service network to improve performance
	///TODO because we won't have to wait for all services to start for checking readiness, but first we have to
	//TODO propagate the Recipes to this layer too and probably move the wait instruction also
	if failedServicesChecks := builtin.allServicesReadinessCheck(ctx, startedAndUpdatedService, parallelism); len(failedServicesChecks) > 0 {
		var allServiceChecksErrMsg string
		for serviceName, serviceErr := range failedServicesChecks {
			serviceMsg := fmt.Sprintf("Service '%v' error:\n%v\n", serviceName, serviceErr)
			allServiceChecksErrMsg = allServiceChecksErrMsg + serviceMsg
		}
		return "", stacktrace.NewError("An error occurred while checking all service, these are the errors by service:\n%s", allServiceChecksErrMsg)
	}
	defer func() {
		if shouldDeleteAllStartedServices {
			builtin.removeAllStartedServices(ctx, startedAndUpdatedService)
		}
	}()

	instructionResult := strings.Builder{}
	instructionResult.WriteString(fmt.Sprintf("Successfully added the following '%d' services:", len(startedServices)))
	for serviceName, serviceObj := range startedAndUpdatedService {
		if err := fillAddServiceReturnValueWithRuntimeValues(serviceObj, builtin.resultUuids[serviceName], builtin.runtimeValueStore); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while adding service return values with result key UUID '%s'", builtin.resultUuids[serviceName])
		}
		instructionResult.WriteString(fmt.Sprintf("\n  Service '%s' added with UUID '%s'", serviceName, serviceObj.GetRegistration().GetUUID()))
	}
	shouldDeleteAllStartedServices = false
	return instructionResult.String(), nil
}

func (builtin *AddServicesCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		for serviceName := range builtin.serviceConfigs {
			enclaveComponents.AddService(serviceName, enclave_structure.ComponentWasLeftIntact)
		}
		return enclave_structure.InstructionIsEqual
	}
	// if other instruction is nil or other instruction is not an add_services instruction, status is unknown
	if other == nil {
		for serviceName := range builtin.serviceConfigs {
			enclaveComponents.AddService(serviceName, enclave_structure.ComponentIsNew)
		}
		return enclave_structure.InstructionIsUnknown
	}

	if other.Type != AddServicesBuiltinName {
		for serviceName := range builtin.serviceConfigs {
			enclaveComponents.AddService(serviceName, enclave_structure.ComponentIsNew)
		}
		return enclave_structure.InstructionIsUnknown
	}

	// The instruction can be re-run only if the set of added services is a superset of what was added by the
	// instruction it's being compared to, so we check that first
	atLeastOneFileHasBeenUpdated := false
	previouslyAddedService := map[service.ServiceName]bool{}
	for _, serviceName := range other.ServiceNames {
		previouslyAddedService[service.ServiceName(serviceName)] = false
	}
	for serviceName, serviceConfig := range builtin.serviceConfigs {
		if _, found := previouslyAddedService[serviceName]; found {
			previouslyAddedService[serviceName] = true // toggle the boolean to true
		}

		// Check whether one file as been updated - if yes the instruction will need to be rerun
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		if filesArtifactsExpansion != nil {
			for _, filesArtifactNames := range filesArtifactsExpansion.ServiceDirpathsToArtifactIdentifiers {
				for _, filesArtifactName := range filesArtifactNames {
					if enclaveComponents.HasFilesArtifactBeenUpdated(filesArtifactName) {
						atLeastOneFileHasBeenUpdated = true
					}
				}
			}
		}
	}
	for _, servicePresentInCurrentInstruction := range previouslyAddedService {
		if !servicePresentInCurrentInstruction {
			// if one service is not present in the current instruction, instruction cannot be re-run
			for serviceName := range builtin.serviceConfigs {
				enclaveComponents.AddService(serviceName, enclave_structure.ComponentIsNew)
			}
			return enclave_structure.InstructionIsUnknown
		}
	}

	if !instructionsAreEqual || atLeastOneFileHasBeenUpdated {
		for serviceName := range builtin.serviceConfigs {
			if _, found := previouslyAddedService[serviceName]; found {
				enclaveComponents.AddService(serviceName, enclave_structure.ComponentIsUpdated)
			} else {
				enclaveComponents.AddService(serviceName, enclave_structure.ComponentIsNew)
			}
		}
		return enclave_structure.InstructionIsUpdate
	}

	// the instruction is equal AND no file have been updated. No need to rerun the instruction
	for serviceName := range builtin.serviceConfigs {
		enclaveComponents.AddService(serviceName, enclave_structure.ComponentWasLeftIntact)
	}
	return enclave_structure.InstructionIsEqual
}

func (builtin *AddServicesCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(AddServicesBuiltinName)
	for serviceName := range builtin.serviceConfigs {
		builder.AddServiceName(serviceName)
	}
}

func (builtin *AddServicesCapabilities) removeAllStartedServices(
	ctx context.Context,
	startedServices map[service.ServiceName]*service.Service,
) {
	//this is not executed with concurrency because the remove service method locks on every call
	for startedServiceName, startedService := range startedServices {
		serviceIdentifier := string(startedService.GetRegistration().GetUUID())
		if _, err := builtin.serviceNetwork.RemoveService(ctx, serviceIdentifier); err != nil {
			logrus.Debugf("Something fails while started all services and we tried to remove all the  created services to rollback the process, but this one '%s' fails throwing this error: '%v', we suggest you to manually remove it", startedServiceName, err)
		}
	}
}

func (builtin *AddServicesCapabilities) allServicesReadinessCheck(
	ctx context.Context,
	startedServices map[service.ServiceName]*service.Service,
	batchSize int,
) map[service.ServiceName]error {
	logrus.Debugf("Checking for all services readiness...")

	concurrencyControlChan := make(chan bool, batchSize)
	defer close(concurrencyControlChan)

	failedServiceChecksSyncMap := &sync.Map{}

	wg := &sync.WaitGroup{}
	for serviceName := range startedServices {
		wg.Add(1)
		// The concurrencyControlChan will block if the buffer is currently full
		concurrencyControlChan <- true
		go builtin.runServiceReadinessCheck(ctx, wg, concurrencyControlChan, serviceName, failedServiceChecksSyncMap)
	}
	wg.Wait()

	failedServiceChecksRegularMap := map[service.ServiceName]error{}

	failedServiceChecksSyncMap.Range(func(serviceNameAny any, serviceErrorAny any) bool {
		if serviceErrorAny != nil {
			serviceName := serviceNameAny.(service.ServiceName)
			serviceError := serviceErrorAny.(error)
			failedServiceChecksRegularMap[serviceName] = serviceError
		}
		return true
	})

	logrus.Debug("All services are ready")

	return failedServiceChecksRegularMap
}

func (builtin *AddServicesCapabilities) Description() string {
	return fmt.Sprintf("Adding '%v' services with names '%v'", len(builtin.serviceConfigs), getNamesAsCommaSeparatedList(builtin.serviceConfigs))
}

func getNamesAsCommaSeparatedList(serviceConfigs map[service.ServiceName]*service.ServiceConfig) string {
	var serviceNames []string
	serviceNameSeparator := ","
	for serviceName := range serviceConfigs {
		serviceNames = append(serviceNames, string(serviceName))
	}
	return strings.Join(serviceNames, serviceNameSeparator)
}

func (builtin *AddServicesCapabilities) runServiceReadinessCheck(
	ctx context.Context,
	wg *sync.WaitGroup,
	concurrencyControlChan chan bool,
	serviceName service.ServiceName,
	failedServiceChecks *sync.Map,
) {
	var serviceErr error
	defer func() {
		failedServiceChecks.Store(serviceName, serviceErr)
		wg.Done()
		//pop a value from the concurrencyControlChan to allow any potentially waiting subroutine to start
		<-concurrencyControlChan
	}()

	readyConditions, found := builtin.readyConditions[serviceName]
	if !found {
		serviceErr = stacktrace.NewError("Expected to find ready conditions for service '%s' in map '%+v', but none was found; this is a bug in Kurtosis", serviceName, builtin.readyConditions)
		return
	}

	if err := runServiceReadinessCheck(
		ctx,
		builtin.serviceNetwork,
		builtin.runtimeValueStore,
		serviceName,
		readyConditions,
	); err != nil {
		serviceErr = stacktrace.Propagate(err, "An error occurred while checking if service '%v' is ready", serviceName)
		return
	}
}

func validateAndConvertConfigsAndReadyConditions(
	serviceNetwork service_network.ServiceNetwork,
	configs starlark.Value,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
) (
	map[service.ServiceName]*service.ServiceConfig,
	map[service.ServiceName]*service_config.ReadyCondition,
	*startosis_errors.InterpretationError,
) {
	configsDict, ok := configs.(*starlark.Dict)
	if !ok {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument should be a dictionary of matching each service name to their respective ServiceConfig object. Got '%s'", ConfigsArgName, reflect.TypeOf(configs))
	}
	if configsDict.Len() == 0 {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument should be a non empty dictionary", ConfigsArgName)
	}
	convertedServiceConfigs := map[service.ServiceName]*service.ServiceConfig{}
	readyConditionsByServiceName := map[service.ServiceName]*service_config.ReadyCondition{}
	for _, serviceName := range configsDict.Keys() {
		serviceNameStr, isServiceNameAString := serviceName.(starlark.String)
		if !isServiceNameAString {
			return nil, nil, startosis_errors.NewInterpretationError("One key of the '%s' dictionary is not a string (was '%s'). Keys of this argument should correspond to service names, which should be strings", ConfigsArgName, reflect.TypeOf(serviceName))
		}

		dictValue, found, err := configsDict.Get(serviceName)
		if err != nil || !found {
			return nil, nil, startosis_errors.NewInterpretationError("Could not extract the value of the '%s' dictionary for key '%s'. This is Kurtosis bug", ConfigsArgName, serviceName)
		}
		serviceConfig, isDictValueAServiceConfig := dictValue.(*service_config.ServiceConfig)
		if !isDictValueAServiceConfig {
			return nil, nil, startosis_errors.NewInterpretationError("One value of the '%s' dictionary is not a ServiceConfig (was '%s'). Values of this argument should correspond to the config of the service to be added", ConfigsArgName, reflect.TypeOf(dictValue))
		}
		apiServiceConfig, interpretationErr := serviceConfig.ToKurtosisType(serviceNetwork, locatorOfModuleInWhichThisBuiltInIsBeingCalled, packageId, packageContentProvider, packageReplaceOptions)
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		convertedServiceConfigs[service.ServiceName(serviceNameStr.GoString())] = apiServiceConfig

		readyConditions, interpretationErr := serviceConfig.GetReadyCondition()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}

		readyConditionsByServiceName[service.ServiceName(serviceNameStr.GoString())] = readyConditions
	}
	return convertedServiceConfigs, readyConditionsByServiceName, nil
}

func makeAndPersistAddServicesInterpretationReturnValue(serviceConfigs map[service.ServiceName]*service.ServiceConfig, runtimeValueStore *runtime_value_store.RuntimeValueStore, interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore) (map[service.ServiceName]string, *starlark.Dict, *startosis_errors.InterpretationError) {
	servicesObjectDict := starlark.NewDict(len(serviceConfigs))
	resultUuids := map[service.ServiceName]string{}
	var err error
	for serviceName, serviceConfig := range serviceConfigs {
		serviceNameStr := starlark.String(serviceName)
		resultUuids[serviceName], err = runtimeValueStore.GetOrCreateValueAssociatedWithService(serviceName)
		if err != nil {
			return nil, nil, startosis_errors.WrapWithInterpretationError(err, "Unable to create runtime value to hold '%v' command return values", AddServicesBuiltinName)
		}
		serviceObject, interpretationErr := makeAddServiceInterpretationReturnValue(serviceNameStr, serviceConfig, resultUuids[serviceName])
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		if err := servicesObjectDict.SetKey(serviceNameStr, serviceObject); err != nil {
			return nil, nil, startosis_errors.WrapWithInterpretationError(err, "Unable to generate the object that should be returned by the '%s' builtin", AddServicesBuiltinName)
		}
		if err = interpretationTimeValueStore.PutService(serviceName, serviceObject); err != nil {
			return nil, nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while persisting the return value for service with name '%v'", serviceName)
		}
	}
	return resultUuids, servicesObjectDict, nil
}

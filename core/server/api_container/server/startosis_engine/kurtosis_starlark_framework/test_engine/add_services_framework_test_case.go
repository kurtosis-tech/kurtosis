package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"go.starlark.net/starlark"
)

type addServicesTestCase struct{}

func (test *addServicesTestCase) GetId() string {
	return add_service.AddServicesBuiltinName
}

func (test *addServicesTestCase) GetInstruction() (*kurtosis_plan_instruction.KurtosisPlanInstruction, error) {
	return add_service.NewAddServices(nil, nil), nil
}

func (test *addServicesTestCase) GetStarlarkCode() (string, error) {
	serviceConfig1 := `ServiceConfig(image="test-image-1", subnetwork="my-subnetwork")`
	serviceConfig2 := `ServiceConfig(image="test-image-2", cpu_allocation=1000, memory_allocation=2048)`
	return fmt.Sprintf(`%s(%s={"service-1": %s, "service-2": %s})`, add_service.AddServicesBuiltinName, add_service.ConfigsArgName, serviceConfig1, serviceConfig2), nil
}

func (test *addServicesTestCase) GetExpectedArguments() (starlark.StringDict, error) {
	configs := starlark.NewDict(2)
	subnetwork := starlark.String("my-subnetwork")
	err := configs.SetKey(
		starlark.String("service-1"),
		kurtosis_types.NewServiceConfig(
			"test-image-1",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			&subnetwork,
			nil,
			nil,
		),
	)
	if err != nil {
		return nil, err
	}

	cpuAllocation := starlark.MakeInt(1000)
	memoryAllocation := starlark.MakeInt(2048)
	err = configs.SetKey(
		starlark.String("service-2"),
		kurtosis_types.NewServiceConfig(
			"test-image-2",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			&cpuAllocation,
			&memoryAllocation,
		),
	)
	if err != nil {
		return nil, err
	}
	return starlark.StringDict{
		add_service.ConfigsArgName: configs,
	}, nil
}

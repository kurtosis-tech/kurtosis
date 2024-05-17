package starlark_run

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
)

type StarlarkRun struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateStarlarkRun *privateStarlarkRun
}

type privateStarlarkRun struct {
	PackageId              string
	SerializedScript       string
	SerializedParams       string
	Parallelism            int32
	RelativePathToMainFile string
	MainFunctionName       string
	ExperimentalFeatures   []int32
	RestartPolicy          int32
	// The params that were used for the very first run of the script
	InitialSerializedParams string
}

func NewStarlarkRun(
	packageId string,
	serializedScript string,
	serializedParams string,
	parallelism int32,
	relativePathToMainFile string,
	mainFunctionName string,
	experimentalFeatures []int32,
	restartPolicy int32,
	initialSerializedParams string,
) *StarlarkRun {
	privateStarlarkRunObj := &privateStarlarkRun{
		PackageId:               packageId,
		SerializedScript:        serializedScript,
		SerializedParams:        serializedParams,
		Parallelism:             parallelism,
		RelativePathToMainFile:  relativePathToMainFile,
		MainFunctionName:        mainFunctionName,
		ExperimentalFeatures:    experimentalFeatures,
		RestartPolicy:           restartPolicy,
		InitialSerializedParams: initialSerializedParams,
	}

	return &StarlarkRun{
		privateStarlarkRun: privateStarlarkRunObj,
	}
}

func (run *StarlarkRun) GetPackageId() string {
	return run.privateStarlarkRun.PackageId
}

func (run *StarlarkRun) SetPackageId(packageId string) {
	run.privateStarlarkRun.PackageId = packageId
}

func (run *StarlarkRun) GetSerializedScript() string {
	return run.privateStarlarkRun.SerializedScript
}

func (run *StarlarkRun) SetSerializedScript(serializedScript string) {
	run.privateStarlarkRun.SerializedScript = serializedScript
}

func (run *StarlarkRun) GetSerializedParams() string {
	return run.privateStarlarkRun.SerializedParams
}

func (run *StarlarkRun) SetSerializedParams(serializedParams string) {
	run.privateStarlarkRun.SerializedParams = serializedParams
}

func (run *StarlarkRun) GetParallelism() int32 {
	return run.privateStarlarkRun.Parallelism
}

func (run *StarlarkRun) SetParallelism(parallelism int32) {
	run.privateStarlarkRun.Parallelism = parallelism
}

func (run *StarlarkRun) GetRelativePathToMainFile() string {
	return run.privateStarlarkRun.RelativePathToMainFile
}

func (run *StarlarkRun) SetRelativePathToMainFile(relativePathToMainFile string) {
	run.privateStarlarkRun.RelativePathToMainFile = relativePathToMainFile
}

func (run *StarlarkRun) GetMainFunctionName() string {
	return run.privateStarlarkRun.MainFunctionName
}

func (run *StarlarkRun) SetMainFunctionName(mainFunctionName string) {
	run.privateStarlarkRun.MainFunctionName = mainFunctionName
}

func (run *StarlarkRun) GetExperimentalFeatures() []int32 {
	return run.privateStarlarkRun.ExperimentalFeatures
}

func (run *StarlarkRun) SetExperimentalFeatures(experimentalFeatures []int32) {
	run.privateStarlarkRun.ExperimentalFeatures = experimentalFeatures
}

func (run *StarlarkRun) GetRestartPolicy() int32 {
	return run.privateStarlarkRun.RestartPolicy
}

func (run *StarlarkRun) SetRestartPolicy(restartPolicy int32) {
	run.privateStarlarkRun.RestartPolicy = restartPolicy
}

func (run *StarlarkRun) GetInitialSerializedParams() string {
	return run.privateStarlarkRun.InitialSerializedParams
}

func (run *StarlarkRun) SetInitialSerializedParams(initialSerializedParams string) {
	run.privateStarlarkRun.InitialSerializedParams = initialSerializedParams
}

func (run *StarlarkRun) MarshalJSON() ([]byte, error) {
	return json.Marshal(run.privateStarlarkRun)
}

func (run *StarlarkRun) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateStarlarkRun{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	run.privateStarlarkRun = unmarshalledPrivateStructPtr
	return nil
}

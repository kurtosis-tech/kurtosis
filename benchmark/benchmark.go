package benchmark

import (
	"fmt"
	"time"
)

type RunBenchmark struct {
	TimeToRun time.Duration
	TimeToCreateEnclave time.Duration
	TimeToUploadStarlarkPackage time.Duration
	TimeToExecuteStarlark time.Duration
}

func (benchmark *RunBenchmark) Print() {
	fmt.Printf("Time to create enclave: %v\n", benchmark.TimeToCreateEnclave)
	fmt.Printf("Time to execute starlark: %v\n", benchmark.TimeToExecuteStarlark)
	fmt.Printf("Time to upload starlark package: %v\n", benchmark.TimeToUploadStarlarkPackage)
	fmt.Printf("Time to run: %v\n", benchmark.TimeToRun)
}

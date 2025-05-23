package benchmark

import (
	"fmt"
	"log"
	"os"
	"time"
)

type RunBenchmark struct {
	TimeToRun                   time.Duration
	TimeToCreateEnclave         time.Duration
	TimeToUploadStarlarkPackage time.Duration
	TimeToExecuteStarlark       time.Duration
}

type APICBenchmark struct {
}

type StartosisBenchmark struct {
	TimeToRunStartosisScript    time.Duration
	TimeToExecuteInstructions   time.Duration
	TimeToValidateInstructions  time.Duration
	TimeToInterpretInstructions time.Duration
}

type KurtosisBackendBenchmark struct {
}

type ServiceNetworkBenchmark struct {
}

func (benchmark *RunBenchmark) Print() {
	fmt.Printf("Time to create enclave: %v\n", benchmark.TimeToCreateEnclave)
	fmt.Printf("Time to execute starlark: %v\n", benchmark.TimeToExecuteStarlark)
	fmt.Printf("Time to upload starlark package: %v\n", benchmark.TimeToUploadStarlarkPackage)
	fmt.Printf("Time to run: %v\n", benchmark.TimeToRun)
}

func (benchmark *StartosisBenchmark) OutputToFile(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("Time to run startosis script: %v\n", benchmark.TimeToRunStartosisScript))
	file.WriteString(fmt.Sprintf("Time to execute instructions: %v\n", benchmark.TimeToExecuteInstructions))
	file.WriteString(fmt.Sprintf("Time to validate instructions: %v\n", benchmark.TimeToValidateInstructions))
	file.WriteString(fmt.Sprintf("Time to interpret instructions: %v\n", benchmark.TimeToInterpretInstructions))
}

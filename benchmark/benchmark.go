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

type KurtosisPlanInstructionBenchmark struct {
	TimeToAddServices              time.Duration
	NumAddServices                 int
	TimeToRunSh                    time.Duration
	NumRunSh                       int
	TimeToRenderTemplates          time.Duration
	NumRenderTemplates             int
	TimeToVerify                   time.Duration
	NumVerify                      int
	TimeToWait                     time.Duration
	NumWait                        int
	TimeToExec                     time.Duration
	NumExec                        int
	TimeToStoreServiceFiles        time.Duration
	NumStoreServiceFiles           int
	TimeToUploadFiles              time.Duration
	NumUploadFiles                 int
	TimeToPrint                    time.Duration
	NumPrint                       int
	TotalTimeExecutingInstructions time.Duration
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

func (benchmark *KurtosisPlanInstructionBenchmark) OutputToFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("Total time executing instructions: %v\n", benchmark.TotalTimeExecutingInstructions))
	file.WriteString(fmt.Sprintf("Time to add services: %v\n", benchmark.TimeToAddServices))
	file.WriteString(fmt.Sprintf("Number of add services: %v\n", benchmark.NumAddServices))
	file.WriteString(fmt.Sprintf("Time to run sh: %v\n", benchmark.TimeToRunSh))
	file.WriteString(fmt.Sprintf("Number of run sh: %v\n", benchmark.NumRunSh))
	file.WriteString(fmt.Sprintf("Time to render templates: %v\n", benchmark.TimeToRenderTemplates))
	file.WriteString(fmt.Sprintf("Number of render templates: %v\n", benchmark.NumRenderTemplates))
	file.WriteString(fmt.Sprintf("Time to verify: %v\n", benchmark.TimeToVerify))
	file.WriteString(fmt.Sprintf("Number of verify: %v\n", benchmark.NumVerify))
	file.WriteString(fmt.Sprintf("Time to wait: %v\n", benchmark.TimeToWait))
	file.WriteString(fmt.Sprintf("Number of wait: %v\n", benchmark.NumWait))
	file.WriteString(fmt.Sprintf("Time to exec: %v\n", benchmark.TimeToExec))
	file.WriteString(fmt.Sprintf("Number of exec: %v\n", benchmark.NumExec))
	file.WriteString(fmt.Sprintf("Time to store service files: %v\n", benchmark.TimeToStoreServiceFiles))
	file.WriteString(fmt.Sprintf("Number of store service files: %v\n", benchmark.NumStoreServiceFiles))
	file.WriteString(fmt.Sprintf("Time to upload files: %v\n", benchmark.TimeToUploadFiles))
	file.WriteString(fmt.Sprintf("Number of upload files: %v\n", benchmark.NumUploadFiles))
	file.WriteString(fmt.Sprintf("Time to print: %v\n", benchmark.TimeToPrint))
	file.WriteString(fmt.Sprintf("Number of print: %v\n", benchmark.NumPrint))
}

package benchmark

import (
	"encoding/csv"
	"fmt"
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
	fmt.Printf("Create enclave: %v\n", benchmark.TimeToCreateEnclave)
	fmt.Printf("Execute starlark: %v\n", benchmark.TimeToExecuteStarlark)
	fmt.Printf("Upload starlark package: %v\n", benchmark.TimeToUploadStarlarkPackage)
	fmt.Printf("Run: %v\n", benchmark.TimeToRun)
}

func (benchmark *RunBenchmark) OutputToFile(filePath string, format string) error {
	return benchmark.outputToCSV()
}

func (benchmark *RunBenchmark) outputToCSV() error {
	filePath := "run_benchmark.csv"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Metric", "Value"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Create enclave", benchmark.TimeToCreateEnclave.String()},
		{"Execute starlark", benchmark.TimeToExecuteStarlark.String()},
		{"Upload starlark package", benchmark.TimeToUploadStarlarkPackage.String()},
		{"Run", benchmark.TimeToRun.String()},
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

func (benchmark *StartosisBenchmark) OutputToFile() error {
	return benchmark.outputToCSV()
}

func (benchmark *StartosisBenchmark) outputToCSV() error {
	filePath := "startosis_benchmark.csv"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Metric", "Value"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Run startosis script", benchmark.TimeToRunStartosisScript.String()},
		{"Execute instructions", benchmark.TimeToExecuteInstructions.String()},
		{"Validate instructions", benchmark.TimeToValidateInstructions.String()},
		{"Interpret instructions", benchmark.TimeToInterpretInstructions.String()},
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

func (benchmark *KurtosisPlanInstructionBenchmark) OutputToFile() error {
	return benchmark.outputToCSV()
}

func (benchmark *KurtosisPlanInstructionBenchmark) outputToCSV() error {
	filePath := "kurtosis_plan_instruction_benchmark.csv"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Metric", "Value", "Count"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Total time executing instructions", benchmark.TotalTimeExecutingInstructions.String(), ""},
		{"Add services", benchmark.TimeToAddServices.String(), fmt.Sprintf("%d", benchmark.NumAddServices)},
		{"Run sh", benchmark.TimeToRunSh.String(), fmt.Sprintf("%d", benchmark.NumRunSh)},
		{"Render templates", benchmark.TimeToRenderTemplates.String(), fmt.Sprintf("%d", benchmark.NumRenderTemplates)},
		{"Verify", benchmark.TimeToVerify.String(), fmt.Sprintf("%d", benchmark.NumVerify)},
		{"Wait", benchmark.TimeToWait.String(), fmt.Sprintf("%d", benchmark.NumWait)},
		{"Exec", benchmark.TimeToExec.String(), fmt.Sprintf("%d", benchmark.NumExec)},
		{"Store service files", benchmark.TimeToStoreServiceFiles.String(), fmt.Sprintf("%d", benchmark.NumStoreServiceFiles)},
		{"Upload files", benchmark.TimeToUploadFiles.String(), fmt.Sprintf("%d", benchmark.NumUploadFiles)},
		{"Print", benchmark.TimeToPrint.String(), fmt.Sprintf("%d", benchmark.NumPrint)},
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

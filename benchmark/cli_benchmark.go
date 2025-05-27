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

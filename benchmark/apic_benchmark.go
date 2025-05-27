package benchmark

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	BenchmarkDataDir = "/run/benchmark-data"
)

type AddServiceBenchmark struct {
	ServiceName               string
	TimeToAddServiceContainer time.Duration
	TimeToReadinessCheck      time.Duration
}

type RunShBenchmark struct {
	TaskName               string
	TimeToAddTaskContainer time.Duration
	TimeToExecWithWait     time.Duration
}

type KurtosisPlanInstructionBenchmark struct {
	TimeToAddServices    time.Duration
	NumAddServices       int
	addServiceBenchmarks []AddServiceBenchmark

	TimeToRunSh     time.Duration
	NumRunSh        int
	runShBenchmarks []RunShBenchmark

	TimeToRenderTemplates time.Duration
	NumRenderTemplates    int

	TimeToVerify time.Duration
	NumVerify    int

	TimeToWait time.Duration
	NumWait    int

	TimeToExec time.Duration
	NumExec    int

	TimeToStoreServiceFiles time.Duration
	NumStoreServiceFiles    int

	TimeToUploadFiles time.Duration
	NumUploadFiles    int

	TimeToPrint time.Duration
	NumPrint    int

	TotalTimeExecutingInstructions time.Duration
}

func NewKurtosisPlanInstructionBenchmark() *KurtosisPlanInstructionBenchmark {
	err := os.MkdirAll(BenchmarkDataDir, 0755)
	if err != nil {
		logrus.Errorf("failed to create benchmark datadirectory: %v", err)
	}
	return &KurtosisPlanInstructionBenchmark{
		addServiceBenchmarks:  make([]AddServiceBenchmark, 0),
		runShBenchmarks:       make([]RunShBenchmark, 0),
		TimeToAddServices:     time.Duration(0),
		NumAddServices:        0,
		TimeToRunSh:           time.Duration(0),
		NumRunSh:              0,
		TimeToRenderTemplates: time.Duration(0),
		NumRenderTemplates:    0,
		TimeToVerify:          time.Duration(0),
	}
}

func (benchmark *KurtosisPlanInstructionBenchmark) AddAddServiceBenchmark(addServiceBenchmark AddServiceBenchmark) {
	benchmark.addServiceBenchmarks = append(benchmark.addServiceBenchmarks, addServiceBenchmark)
}

func (benchmark *KurtosisPlanInstructionBenchmark) AddRunShBenchmark(runShBenchmark RunShBenchmark) {
	benchmark.runShBenchmarks = append(benchmark.runShBenchmarks, runShBenchmark)
}

func (benchmark *KurtosisPlanInstructionBenchmark) OutputToFile() error {
	if err := benchmark.outputAddServicesBenchmarksToCsv(); err != nil {
		return err
	}
	if err := benchmark.outputRunShBenchmarksToCsv(); err != nil {
		return err
	}
	return benchmark.outputToCSV()
}

func (benchmark *KurtosisPlanInstructionBenchmark) outputToCSV() error {
	filePath := fmt.Sprintf("%s/kurtosis_plan_instruction_benchmark.csv", BenchmarkDataDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Instruction Name", "Total Time in Instruction", "Number of Instructions"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Total time executing instructions", benchmark.TotalTimeExecutingInstructions.String(), fmt.Sprintf("%d", 1)},
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

func (benchmark *KurtosisPlanInstructionBenchmark) outputRunShBenchmarksToCsv() error {
	filePath := fmt.Sprintf("%s/run_sh_benchmark.csv", BenchmarkDataDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Task Name", "Time To Add Task Container", "Time To Exec With Wait"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	records := [][]string{}
	for _, b := range benchmark.runShBenchmarks {
		records = append(records, []string{b.TaskName, b.TimeToAddTaskContainer.String(), b.TimeToExecWithWait.String()})
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

func (benchmark *KurtosisPlanInstructionBenchmark) outputAddServicesBenchmarksToCsv() error {
	filePath := fmt.Sprintf("%s/add_services_benchmark.csv", BenchmarkDataDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Service Name", "Time To Add Service Container", "Time To Readiness Check"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	records := [][]string{}
	for _, b := range benchmark.addServiceBenchmarks {
		records = append(records, []string{b.ServiceName, b.TimeToAddServiceContainer.String(), b.TimeToReadinessCheck.String()})
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

type StartosisBenchmark struct {
	TimeToRunStartosisScript    time.Duration
	TimeToExecuteInstructions   time.Duration
	TimeToValidateInstructions  time.Duration
	TimeToInterpretInstructions time.Duration
}

func (benchmark *StartosisBenchmark) OutputToFile() error {
	return benchmark.outputToCSV()
}

func (benchmark *StartosisBenchmark) outputToCSV() error {
	filePath := fmt.Sprintf("%s/startosis_benchmark.csv", BenchmarkDataDir)
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

type APICBenchmark struct {
}

type ServiceNetworkBenchmark struct {
}

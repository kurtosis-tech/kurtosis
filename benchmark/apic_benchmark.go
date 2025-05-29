package benchmark

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
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

type RequestBenchmark struct {
	ServiceName   string
	TimeToRequest time.Duration
}

type KurtosisPlanInstructionBenchmark struct {
	TimeToAddServices    time.Duration
	NumAddServices       int
	addServiceBenchmarks []AddServiceBenchmark

	TimeToRunSh     time.Duration
	NumRunSh        int
	runShBenchmarks []RunShBenchmark

	TimeToRequest     time.Duration
	NumRequest        int
	requestBenchmarks []RequestBenchmark

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
		logrus.Errorf("failed to create benchmark data directory: %v", err)
	}
	return &KurtosisPlanInstructionBenchmark{
		addServiceBenchmarks:           make([]AddServiceBenchmark, 0),
		runShBenchmarks:                make([]RunShBenchmark, 0),
		TimeToAddServices:              time.Duration(0),
		NumAddServices:                 0,
		TimeToRunSh:                    time.Duration(0),
		NumRunSh:                       0,
		TimeToRenderTemplates:          time.Duration(0),
		NumRenderTemplates:             0,
		TimeToVerify:                   time.Duration(0),
		TimeToRequest:                  time.Duration(0),
		NumRequest:                     0,
		requestBenchmarks:              make([]RequestBenchmark, 0),
		TimeToStoreServiceFiles:        time.Duration(0),
		NumStoreServiceFiles:           0,
		TimeToUploadFiles:              time.Duration(0),
		NumUploadFiles:                 0,
		TimeToPrint:                    time.Duration(0),
		NumPrint:                       0,
		TotalTimeExecutingInstructions: time.Duration(0),
	}
}

func (benchmark *KurtosisPlanInstructionBenchmark) AddAddServiceBenchmark(addServiceBenchmark AddServiceBenchmark) {
	benchmark.addServiceBenchmarks = append(benchmark.addServiceBenchmarks, addServiceBenchmark)
}

func (benchmark *KurtosisPlanInstructionBenchmark) AddRunShBenchmark(runShBenchmark RunShBenchmark) {
	benchmark.runShBenchmarks = append(benchmark.runShBenchmarks, runShBenchmark)
}

func (benchmark *KurtosisPlanInstructionBenchmark) AddRequestBenchmark(requestBenchmark RequestBenchmark) {
	benchmark.requestBenchmarks = append(benchmark.requestBenchmarks, requestBenchmark)
}

func (benchmark *KurtosisPlanInstructionBenchmark) OutputToFile() error {
	if err := benchmark.outputAddServicesBenchmarksToCsv(); err != nil {
		return err
	}
	if err := benchmark.outputRunShBenchmarksToCsv(); err != nil {
		return err
	}
	if err := benchmark.outputRequestBenchmarksToCsv(); err != nil {
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
	if err := writer.Write([]string{"Instruction Name", "Total Time in Instruction (s)", "Number of Instructions"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Total time executing instructions", durationToSeconds(benchmark.TotalTimeExecutingInstructions), fmt.Sprintf("%d", 1)},
		{"Add services", durationToSeconds(benchmark.TimeToAddServices), fmt.Sprintf("%d", benchmark.NumAddServices)},
		{"Run sh", durationToSeconds(benchmark.TimeToRunSh), fmt.Sprintf("%d", benchmark.NumRunSh)},
		{"Render templates", durationToSeconds(benchmark.TimeToRenderTemplates), fmt.Sprintf("%d", benchmark.NumRenderTemplates)},
		{"Verify", durationToSeconds(benchmark.TimeToVerify), fmt.Sprintf("%d", benchmark.NumVerify)},
		{"Wait", durationToSeconds(benchmark.TimeToWait), fmt.Sprintf("%d", benchmark.NumWait)},
		{"Exec", durationToSeconds(benchmark.TimeToExec), fmt.Sprintf("%d", benchmark.NumExec)},
		{"Store service files", durationToSeconds(benchmark.TimeToStoreServiceFiles), fmt.Sprintf("%d", benchmark.NumStoreServiceFiles)},
		{"Upload files", durationToSeconds(benchmark.TimeToUploadFiles), fmt.Sprintf("%d", benchmark.NumUploadFiles)},
		{"Print", durationToSeconds(benchmark.TimeToPrint), fmt.Sprintf("%d", benchmark.NumPrint)},
		{"Request", durationToSeconds(benchmark.TimeToRequest), fmt.Sprintf("%d", benchmark.NumRequest)},
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

	if err := writer.Write([]string{"Task Name", "Time To Add Task Container (s)", "Time To Exec With Wait (s)"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	records := [][]string{}
	for _, b := range benchmark.runShBenchmarks {
		records = append(records, []string{b.TaskName, durationToSeconds(b.TimeToAddTaskContainer), durationToSeconds(b.TimeToExecWithWait)})
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
	if err := writer.Write([]string{"Service Name", "Time To Add Service Container (s)", "Time To Readiness Check (s)"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	records := [][]string{}
	for _, b := range benchmark.addServiceBenchmarks {
		records = append(records, []string{b.ServiceName, durationToSeconds(b.TimeToAddServiceContainer), durationToSeconds(b.TimeToReadinessCheck)})
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

func (benchmark *KurtosisPlanInstructionBenchmark) outputRequestBenchmarksToCsv() error {
	filePath := fmt.Sprintf("%s/request_benchmark.csv", BenchmarkDataDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Service Name", "Time To Request (s)"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	records := [][]string{}
	for _, b := range benchmark.requestBenchmarks {
		records = append(records, []string{b.ServiceName, durationToSeconds(b.TimeToRequest)})
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
	if err := writer.Write([]string{"Metric", "Value (s)"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	records := [][]string{
		{"Run startosis script", durationToSeconds(benchmark.TimeToRunStartosisScript)},
		{"Execute instructions", durationToSeconds(benchmark.TimeToExecuteInstructions)},
		{"Validate instructions", durationToSeconds(benchmark.TimeToValidateInstructions)},
		{"Interpret instructions", durationToSeconds(benchmark.TimeToInterpretInstructions)},
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %v", err)
	}

	return nil
}

// durationToSeconds converts a time.Duration to a string representation in seconds
func durationToSeconds(d time.Duration) string {
	return fmt.Sprintf("%.6f", d.Seconds())
}

type APICBenchmark struct {
}

type ServiceNetworkBenchmark struct {
}

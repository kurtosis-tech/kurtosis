package benchmark

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type KurtosisBackendBenchmark struct {
	RegisterUserServicesBenchmark       *RegisterUserServicesBenchmark
	StartUserServicesBenchmark          *StartUserServicesBenchmark
	RunUserServiceExecCommandsBenchmark *RunUserServiceExecCommandBenchmark
}

type RegisterUserService struct {
	Name     string
	Duration time.Duration
}

// output register user services benchmark to csv
type RegisterUserServicesBenchmark struct {
	TimeToRegisterUserServices     time.Duration
	NumTimesToRegisterUserServices int

	RegisterUserServices []RegisterUserService
}

type StartUserService struct {
	Name     string
	Duration time.Duration
}

// output start user services benchmark to csv
type StartUserServicesBenchmark struct {
	TimeToStartUserServices     time.Duration
	NumTimesToStartUserServices int

	StartUserServices []StartUserService
}

type UserServiceExecCommand struct {
	Name     string
	Duration time.Duration
}

// output run user service exec commands benchmark to csv
type RunUserServiceExecCommandBenchmark struct {
	TimeToRunUserServiceExecCommands     time.Duration
	NumTimesToRunUserServiceExecCommands int

	UserServiceExecCommands []UserServiceExecCommand
}

func NewKurtosisBackendBenchmark() *KurtosisBackendBenchmark {
	err := os.MkdirAll(BenchmarkDataDir, 0755)
	if err != nil {
		logrus.Errorf("failed to create benchmark data directory: %v", err)
	}
	return &KurtosisBackendBenchmark{
		RegisterUserServicesBenchmark: &RegisterUserServicesBenchmark{
			TimeToRegisterUserServices:     time.Duration(0),
			NumTimesToRegisterUserServices: 0,
			RegisterUserServices:           []RegisterUserService{},
		},
		StartUserServicesBenchmark: &StartUserServicesBenchmark{
			TimeToStartUserServices:     time.Duration(0),
			NumTimesToStartUserServices: 0,
			StartUserServices:           []StartUserService{},
		},
		RunUserServiceExecCommandsBenchmark: &RunUserServiceExecCommandBenchmark{
			TimeToRunUserServiceExecCommands:     time.Duration(0),
			NumTimesToRunUserServiceExecCommands: 0,
			UserServiceExecCommands:              []UserServiceExecCommand{},
		},
	}
}

func (benchmark *KurtosisBackendBenchmark) AddRegisterUserService(registerUserService RegisterUserService) {
	benchmark.RegisterUserServicesBenchmark.RegisterUserServices = append(benchmark.RegisterUserServicesBenchmark.RegisterUserServices, registerUserService)
}

func (benchmark *KurtosisBackendBenchmark) AddStartUserService(startUserService StartUserService) {
	benchmark.StartUserServicesBenchmark.StartUserServices = append(benchmark.StartUserServicesBenchmark.StartUserServices, startUserService)
}

func (benchmark *KurtosisBackendBenchmark) AddUserServiceExecCommand(userServiceExecCommand UserServiceExecCommand) {
	benchmark.RunUserServiceExecCommandsBenchmark.UserServiceExecCommands = append(benchmark.RunUserServiceExecCommandsBenchmark.UserServiceExecCommands, userServiceExecCommand)
}

func (benchmark *KurtosisBackendBenchmark) OutputToFile() error {
	if err := benchmark.outputKurtosisBackendBenchmarkToCSV(); err != nil {
		return fmt.Errorf("failed to output Kurtosis backend benchmark to CSV: %v", err)
	}
	return nil
}

func (benchmark *KurtosisBackendBenchmark) outputKurtosisBackendBenchmarkToCSV() error {
	filePath := fmt.Sprintf("%s/kurtosis_backend_benchmark.csv", BenchmarkDataDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Kurtosis Backend Operation", "Total Time in Operation (s)", "Number of Operations"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}
	records := [][]string{
		{"Register User Services", durationToSeconds(benchmark.RegisterUserServicesBenchmark.TimeToRegisterUserServices), fmt.Sprintf("%d", benchmark.RegisterUserServicesBenchmark.NumTimesToRegisterUserServices)},
		{"Start User Services", durationToSeconds(benchmark.StartUserServicesBenchmark.TimeToStartUserServices), fmt.Sprintf("%d", benchmark.StartUserServicesBenchmark.NumTimesToStartUserServices)},
		{"Run User Service Exec Commands", durationToSeconds(benchmark.RunUserServiceExecCommandsBenchmark.TimeToRunUserServiceExecCommands), fmt.Sprintf("%d", benchmark.RunUserServiceExecCommandsBenchmark.NumTimesToRunUserServiceExecCommands)},
	}

	writer.WriteAll(records)
	return nil
}

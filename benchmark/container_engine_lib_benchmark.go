package benchmark

import "time"

type KurtosisBackendBenchmark struct {
	TimeToCreateEnclave   time.Duration
	TimeToRegisterService time.Duration
	TimeToStartService    time.Duration
}

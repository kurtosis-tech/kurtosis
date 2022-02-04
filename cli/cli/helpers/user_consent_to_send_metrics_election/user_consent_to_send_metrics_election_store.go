package user_consent_to_send_metrics_election

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"sync"
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentUserConsentToSendMetricsElectionStore *UserConsentToSendMetricsElectionStore
	once                                         sync.Once
)

type UserConsentToSendMetricsElectionStore struct {
	mutex *sync.RWMutex
}

func GetUserConsentToSendMetricsElectionStore() *UserConsentToSendMetricsElectionStore {
	// NOTE: We use a 'once' to initialize the UserConsentToSendMetricsElectionStore because it contains a mutex to guard
	//the file, and we don't ever want multiple UserConsentToSendMetricsElectionStore instances in existence
	once.Do(func() {
		currentUserConsentToSendMetricsElectionStore = &UserConsentToSendMetricsElectionStore{mutex: &sync.RWMutex{}}
	})
	return currentUserConsentToSendMetricsElectionStore
}

func (store *UserConsentToSendMetricsElectionStore) Exist() (bool, error) {
	filepath, err := host_machine_directories.GetUserConsentToSendMetricsElectionFilepath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the user consent to send metrics election filepath")
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred checking if filepath '%v' exists", filepath)
	}
	return true, nil
}

func (store *UserConsentToSendMetricsElectionStore) Create() error {

	filepath, err := host_machine_directories.GetUserConsentToSendMetricsElectionFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the user consent to send metrics election filepath")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating file with filepath '%v'", filepath)
	}
	file.Close()

	return nil
}

// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func (store *UserConsentToSendMetricsElectionStore) remove() error {
	filepath, err := host_machine_directories.GetUserConsentToSendMetricsElectionFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the user consent to send metrics election filepath")
	}

	if err := os.Remove(filepath); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing file with filepath '%v'", filepath)
	}

	return nil
}

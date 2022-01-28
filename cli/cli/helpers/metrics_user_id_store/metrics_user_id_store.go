package metrics_user_id_store

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//DO NOT CHANGE THIS VALUE
	//Changing this value will change the user IDs that get generated
	//which will truncate our ability to analyze user historical trends
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	applicationID = "kurtosis"

	metricsUserIDFilePermissions os.FileMode = 0644
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentMetricsUserIDStore *MetricsUserIDStore
	once sync.Once
)

type MetricsUserIDStore struct {
	mutex *sync.RWMutex
}

func GetMetricsUserIDStore() *MetricsUserIDStore {
	// NOTE: We use a 'once' to initialize the MetricsUserIDStore because it contains a mutex to guard
	//the file, and we don't ever want multiple MetricsUserIDStore instances in existence
	once.Do(func() {
		currentMetricsUserIDStore = &MetricsUserIDStore{mutex: &sync.RWMutex{}}
	})
	return currentMetricsUserIDStore
}

func (store *MetricsUserIDStore) GetUserID() (string, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	shouldCreateNewUserId := false

	userID, err := store.getMetricsUserIDFromFile()
	if err != nil {
		if os.IsNotExist(err) {
			shouldCreateNewUserId = true
		} else {
			return "", stacktrace.Propagate(err, "An error occurred getting metrics user id from file")
		}
	}

	if userID == "" {
		shouldCreateNewUserId = true
	}

	if shouldCreateNewUserId {
		userID, err = machineid.ProtectedID(applicationID)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred generating anonimazed user ID")
		}
		if err = store.saveMetricsUserIdFile(userID); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred saving metrics user id in file")
		}
	}

	return userID, nil
}

// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func (store *MetricsUserIDStore) getMetricsUserIDFromFile() (string, error) {
	filepath, err := host_machine_directories.GetMetricsUserIdFilepath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath")
	}
	logrus.Debugf("Metrics user id filepath: '%v'", filepath)

	fileContent, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := fileContent.Close(); err != nil {
			logrus.Warnf("We tried to close the metrics user id file, but doing so threw an error:\n%v", err)
		}
	}()

	fileContentBytes, err := ioutil.ReadAll(fileContent)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading metrics user id file")
	}

	fileContentStr := string(fileContentBytes)

	return fileContentStr, nil
}

func (store *MetricsUserIDStore) saveMetricsUserIdFile(metricsUserId string) error {

	fileContent := append([]byte{}, metricsUserId...)

	logrus.Debugf("Saving metrics user id in file...")

	filepath, err := host_machine_directories.GetMetricsUserIdFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath")
	}

	err = ioutil.WriteFile(filepath, fileContent, metricsUserIDFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing metrics user id file '%v'", filepath)
	}
	logrus.Debugf("Metrics user id file saved")
	return nil
}

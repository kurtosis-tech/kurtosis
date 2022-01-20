package metrics_tracker

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	applicationID = "kurtosis-cli"

	metricsUserIDFilePermissions os.FileMode = 0644
)

type yamlContent struct {
	MetricsUserID string `yaml:"metrics-user-id"`
}

type MetricsUserIDStore struct {
	mutex *sync.Mutex
}

func NewMetricsUserIDStore() *MetricsUserIDStore {
	return &MetricsUserIDStore{mutex: &sync.Mutex{}}
}

func (store *MetricsUserIDStore) GetUserID() (string, error) {

	userID, err := store.getMetricsUserIDFromYAMLFile()
	if err != nil {
		if os.IsNotExist(err) {
			userID, err = machineid.ProtectedID(applicationID)
			if err != nil {
				return "", stacktrace.Propagate(err, "An error occurred generating protected machine ID")
			}
			if err = store.saveMetricsUserIdYAMLFile(userID); err != nil {
				return "", stacktrace.Propagate(err, "An error occurred saving metrics user id in YAML file")
			}
		} else {
			return "", stacktrace.Propagate(err, "An error occurred getting metrics user id from YAML file")
		}
	}

	return userID, nil
}

// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func (store *MetricsUserIDStore) getMetricsUserIDFromYAMLFile() (string, error) {
	filepath, err := host_machine_directories.GetMetricsUserIdYAMLFilepath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the metrics user id YAML filepath")
	}
	logrus.Debugf("Metrics user id YAML filepath: '%v'", filepath)

	yamlFile, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := yamlFile.Close(); err != nil {
			logrus.Warnf("We tried to close the metrics user id YAML file, but doing so threw an error:\n%v", err)
		}
	}()

	fileContentBytes, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading metrics user id YAML file")
	}

	newYAMLContent := &yamlContent{}

	if err := yaml.Unmarshal(fileContentBytes, newYAMLContent); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred unmarshalling metrics user id YAML file content '%v'", fileContentBytes)
	}

	return newYAMLContent.MetricsUserID, nil
}

func (store *MetricsUserIDStore) saveMetricsUserIdYAMLFile(metricsUserId string) error {

	newYAMLContent := &yamlContent{}
	newYAMLContent.MetricsUserID = metricsUserId

	marshalledYAMLContent, err := yaml.Marshal(newYAMLContent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling metrics user id content '%+v' to a YAML content", newYAMLContent)
	}

	logrus.Debugf("Saving metrics user id in YAML file...")

	filepath, err := host_machine_directories.GetMetricsUserIdYAMLFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the metrics user id YAML filepath")
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	err = ioutil.WriteFile(filepath, marshalledYAMLContent, metricsUserIDFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing metrics user id YAML file '%v'", filepath)
	}
	logrus.Debugf("Metrics user id YAML file saved")
	return nil
}

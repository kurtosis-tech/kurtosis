package user_metrics_election_event_backlog

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

const (
	filePermissions os.FileMode = 0644
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentUserMetricsElectionEventBacklog *userMetricsElectionEventBacklog
	once                                   sync.Once
)

type userMetricsElectionEventBacklog struct {
	mutex    *sync.RWMutex
}

func GetUserMetricsElectionEventBacklog() *userMetricsElectionEventBacklog {
	// NOTE: We use a 'once' to initialize the userMetricsElectionEventBacklog because it contains a mutex to guard
	//the file, and we don't ever want multiple userMetricsElectionEventBacklog instances in existence
	once.Do(func() {
		currentUserMetricsElectionEventBacklog = &userMetricsElectionEventBacklog{
			mutex:    &sync.RWMutex{},
		}
	})
	return currentUserMetricsElectionEventBacklog
}

// Sets the backlog contents, creating the file if it doesn't exist and overwriting it if it does
func (eventBacklog *userMetricsElectionEventBacklog) Set(shouldSendMetrics bool) error {
	eventBacklog.mutex.Lock()
	defer eventBacklog.mutex.Unlock()

	shouldSendMetricsStr := strconv.FormatBool(shouldSendMetrics)
	fileContent := []byte(shouldSendMetricsStr)

	filepath, err := host_machine_directories.GetUserSendMetricsElectionFilepath()
	if err != nil {
		logrus.Debugf( "An error occurred getting the user-send-metrics-election filepath\n%v", err)
	}

	if err := ioutil.WriteFile(filepath, fileContent, filePermissions); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing file '%v'", filepath)
	}

	return nil
}

// Gets the file contents if it exists, and if not returns false for `hasValue`
func (eventBacklog *userMetricsElectionEventBacklog) Get() (shouldSendMetrics bool, hasValue bool, err error) {
	eventBacklog.mutex.RLock()
	defer eventBacklog.mutex.RUnlock()

	filepath, err := host_machine_directories.GetUserSendMetricsElectionFilepath()
	if err != nil {
		logrus.Debugf( "An error occurred getting the user-send-metrics-election filepath\n%v", err)
	}

	fileContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, stacktrace.Propagate(err, "An error occurred reading file with filepath '%v'", filepath)
	}

	shouldSendMetricsStr := string(fileContent)
	shouldSendMetrics, err = strconv.ParseBool(shouldSendMetricsStr)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "An error occurred parsing string '%v' to boolean", shouldSendMetricsStr)
	}

	return shouldSendMetrics, true, nil
}

// If the file exists, deletes it; if not, does nothing
func (eventBacklog *userMetricsElectionEventBacklog) Clear() error {

	filepath, err := host_machine_directories.GetUserSendMetricsElectionFilepath()
	if err != nil {
		logrus.Debugf( "An error occurred getting the user-send-metrics-election filepath\n%v", err)
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return stacktrace.Propagate(err, "An error occurred checking if file with filepath '%v' exists", filepath)
	}

	if err := os.Remove(filepath); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing file with filepath '%v'", filepath)
	}
	return nil
}

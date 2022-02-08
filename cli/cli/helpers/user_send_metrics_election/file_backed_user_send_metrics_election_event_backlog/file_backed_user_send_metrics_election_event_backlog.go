package file_backed_user_send_metrics_election_event_backlog

import (
	"github.com/kurtosis-tech/stacktrace"
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
	currentFileBackedUserSendMetricsElectionEventBacklog *fileBackedUserSendMetricsElectionEventBacklog
	once                                                 sync.Once
)

type fileBackedUserSendMetricsElectionEventBacklog struct {
	mutex *sync.RWMutex
	fileBackedFilepath string
}

func GetFileBackedUserSendMetricsElectionEventBacklog(fileBackedFilepath string) *fileBackedUserSendMetricsElectionEventBacklog {
	// NOTE: We use a 'once' to initialize the fileBackedUserSendMetricsElectionEventBacklog because it contains a mutex to guard
	//the file, and we don't ever want multiple fileBackedUserSendMetricsElectionEventBacklog instances in existence
	once.Do(func() {
		currentFileBackedUserSendMetricsElectionEventBacklog = &fileBackedUserSendMetricsElectionEventBacklog{
			mutex: &sync.RWMutex{},
			fileBackedFilepath: fileBackedFilepath,
		}
	})
	return currentFileBackedUserSendMetricsElectionEventBacklog
}

// Sets the backlog contents, creating the file if it doesn't exist and overwriting it if it does
func (eventBacklog *fileBackedUserSendMetricsElectionEventBacklog) Set(shouldSendMetrics bool) error {
	eventBacklog.mutex.Lock()
	defer eventBacklog.mutex.Unlock()

	shouldSendMetricsStr := strconv.FormatBool(shouldSendMetrics)
	fileContent := []byte(shouldSendMetricsStr)

	if err := ioutil.WriteFile(eventBacklog.fileBackedFilepath, fileContent, filePermissions); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing file '%v'", eventBacklog.fileBackedFilepath)
	}

	return nil
}

// Gets the file contents if it exists, and if not returns false for `hasValue`
func (eventBacklog *fileBackedUserSendMetricsElectionEventBacklog) Get() (shouldSendMetrics bool, hasValue bool, err error) {
	eventBacklog.mutex.RLock()
	defer eventBacklog.mutex.RUnlock()

	fileContent, err := ioutil.ReadFile(eventBacklog.fileBackedFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, stacktrace.Propagate(err, "An error occurred reading file with filepath '%v'", eventBacklog.fileBackedFilepath)
	}

	shouldSendMetricsStr := string(fileContent)
	shouldSendMetrics, err = strconv.ParseBool(shouldSendMetricsStr)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "An error occurred parsing string '%' to boolean", shouldSendMetricsStr)
	}

	return shouldSendMetrics, true, nil
}

// If the file exists, deletes it; if not, does nothing
func (eventBacklog *fileBackedUserSendMetricsElectionEventBacklog) Clear() error {
	if err := os.Remove(eventBacklog.fileBackedFilepath); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing file with filepath '%v'", eventBacklog.fileBackedFilepath)
	}
	return nil
}

package persistence

import (
	"github.com/adrg/xdg"
	"github.com/kurtosis-tech/kurtosis/contexts-state-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"path"
	"sync"
)

const (
	applicationDirname = "kurtosis" // TODO: this is common to Kurtosis config, should refactor

	contextStateFileName = "contexts-state.json"
	defaultFilePerm      = 0644
)

var (
	serializer = protojson.MarshalOptions{
		Multiline: true,
	}
)

type FileBackedConfigPersistence struct {
	*sync.RWMutex

	backingFilePath string
}

func NewFileBackedConfigPersistence() *FileBackedConfigPersistence {
	return &FileBackedConfigPersistence{
		RWMutex:         &sync.RWMutex{},
		backingFilePath: "",
	}
}

func newFileBackedConfigPersistenceForTesting(customFilePath string) *FileBackedConfigPersistence {
	return &FileBackedConfigPersistence{
		RWMutex:         &sync.RWMutex{},
		backingFilePath: customFilePath,
	}
}

func (persistence *FileBackedConfigPersistence) PersistContextsState(newContextsState *generated.KurtosisContextsState) error {
	if err := persistence.init(); err != nil {
		return stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}
	return persistence.persistContextsStateInternal(newContextsState)
}

func (persistence *FileBackedConfigPersistence) LoadContextsState() (*generated.KurtosisContextsState, error) {
	if err := persistence.init(); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}

	persistence.RLock()
	defer persistence.RUnlock()

	contextsStateFileContent, err := os.ReadFile(persistence.backingFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read context config file at '%s'",
			persistence.backingFilePath)
	}
	contextsState := &generated.KurtosisContextsState{}
	if err = protojson.Unmarshal(contextsStateFileContent, contextsState); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to deserialize content of context config file at '%s'",
			persistence.backingFilePath)
	}
	return contextsState, nil
}

func (persistence *FileBackedConfigPersistence) init() error {
	if persistence.backingFilePath != "" {
		// already initialized
		return nil
	}

	contextsStateFilePath, err := getContextsSTateFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to get file path to store config to")
	}

	if _, err = os.Stat(contextsStateFilePath); err != nil {
		if !os.IsNotExist(err) {
			return stacktrace.Propagate(err, "Unexpected error checking if context config file exists at '%s'",
				contextsStateFilePath)
		}
	} else {
		// file already exists, no need to create it
		persistence.backingFilePath = contextsStateFilePath
		return nil
	}

	persistence.backingFilePath = contextsStateFilePath
	if err = persistence.persistContextsStateInternal(defaultContextsState); err != nil {
		return stacktrace.Propagate(err, "Failed to write default contexts config to file")
	}
	return nil
}

func (persistence *FileBackedConfigPersistence) persistContextsStateInternal(newContextsState *generated.KurtosisContextsState) error {
	persistence.Lock()
	defer persistence.Unlock()

	serializedConfig, err := serializer.Marshal(newContextsState)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to serialize content of contexts config object to JSON")
	}
	if err = os.WriteFile(persistence.backingFilePath, serializedConfig, defaultFilePerm); err != nil {
		return stacktrace.Propagate(err, "Unable to write context config file to '%s'",
			persistence.backingFilePath)
	}
	return nil
}

func getContextsSTateFilePath() (string, error) {
	contextConfigFilePath, err := xdg.ConfigFile(path.Join(applicationDirname, contextStateFileName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to get contexts state file path")
	}
	return contextConfigFilePath, nil
}

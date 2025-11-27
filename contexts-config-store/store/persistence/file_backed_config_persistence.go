package persistence

import (
	"github.com/adrg/xdg"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store/serde"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
	"sync"
)

const (
	applicationDirname = "kurtosis" // TODO: this is common to Kurtosis config, should refactor

	contextConfigFileName = "contexts-config.json"
	defaultFilePerm       = 0644
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

func (persistence *FileBackedConfigPersistence) PersistContextsConfig(newContextsConfig *generated.KurtosisContextsConfig) error {
	if err := persistence.init(); err != nil {
		return stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}
	return persistence.persistContextsConfigInternal(newContextsConfig)
}

func (persistence *FileBackedConfigPersistence) LoadContextsConfig() (*generated.KurtosisContextsConfig, error) {
	if err := persistence.init(); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}

	persistence.RLock()
	defer persistence.RUnlock()

	contextsConfigFileContent, err := os.ReadFile(persistence.backingFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read context config file at '%s'",
			persistence.backingFilePath)
	}
	contextsConfig, err := serde.DeserializeKurtosisContextsConfig(contextsConfigFileContent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to deserialize content of context config file at '%s'",
			persistence.backingFilePath)
	}
	return contextsConfig, nil
}

func (persistence *FileBackedConfigPersistence) init() error {
	if persistence.backingFilePath != "" {
		// already initialized
		return nil
	}

	contextsConfigFilePath, err := getContextsConfigFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to get file path to store config to")
	}

	if _, err = os.Stat(contextsConfigFilePath); err != nil {
		if !os.IsNotExist(err) {
			return stacktrace.Propagate(err, "Unexpected error checking if context config file exists at '%s'",
				contextsConfigFilePath)
		}
	} else {
		// file already exists, no need to create it
		persistence.backingFilePath = contextsConfigFilePath
		return nil
	}

	persistence.backingFilePath = contextsConfigFilePath
	newDefaultContextsConfig, err := NewDefaultContextsConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to generate a new default contexts config")
	}
	if err = persistence.persistContextsConfigInternal(newDefaultContextsConfig); err != nil {
		return stacktrace.Propagate(err, "Failed to write default contexts config to file")
	}
	return nil
}

func (persistence *FileBackedConfigPersistence) persistContextsConfigInternal(newContextsConfig *generated.KurtosisContextsConfig) error {
	persistence.Lock()
	defer persistence.Unlock()

	serializedContextsConfig, err := serde.SerializeKurtosisContextsConfig(newContextsConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to serialize content of contexts config object to JSON")
	}
	if err = os.WriteFile(persistence.backingFilePath, serializedContextsConfig, defaultFilePerm); err != nil {
		return stacktrace.Propagate(err, "Unable to write context config file to '%s'",
			persistence.backingFilePath)
	}
	return nil
}

func getContextsConfigFilePath() (string, error) {
	contextConfigFilePath, err := xdg.ConfigFile(path.Join(applicationDirname, contextConfigFileName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to get contexts config file path")
	}
	return contextConfigFilePath, nil
}

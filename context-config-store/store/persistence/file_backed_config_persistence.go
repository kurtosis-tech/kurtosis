package persistence

import (
	"github.com/adrg/xdg"
	"github.com/kurtosis-tech/kurtosis/context-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"path"
	"sync"
)

const (
	applicationDirname = "kurtosis" // TODO: this is common to Kurtosis config, should refactor

	fileName        = "context-config.json"
	defaultFilePerm = 0644
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

func (persistence *FileBackedConfigPersistence) PersistContextConfig(contextConfig *generated.KurtosisContextsConfig) error {
	if err := persistence.init(); err != nil {
		return stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}

	persistence.Lock()
	defer persistence.Unlock()

	return persistence.persistContextConfigInternal(contextConfig)
}

func (persistence *FileBackedConfigPersistence) LoadContextConfig() (*generated.KurtosisContextsConfig, error) {
	if err := persistence.init(); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to initialize context config persistence")
	}

	persistence.RLock()
	defer persistence.RUnlock()

	contextConfigFileContent, err := os.ReadFile(persistence.backingFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read context config file at '%s'",
			persistence.backingFilePath)
	}
	contextConfig := &generated.KurtosisContextsConfig{}
	if err = protojson.Unmarshal(contextConfigFileContent, contextConfig); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to deserialize content of context config file at '%s'",
			persistence.backingFilePath)
	}
	return contextConfig, nil
}

func (persistence *FileBackedConfigPersistence) init() error {
	if persistence.backingFilePath != "" {
		// already initialized
		return nil
	}

	configFilePath, err := getContextConfigFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to get file path to store config to")
	}

	if _, err = os.Stat(configFilePath); err != nil {
		if !os.IsNotExist(err) {
			return stacktrace.Propagate(err, "Unexpected error checking if context config file exists at '%s'",
				configFilePath)
		}
	} else {
		// file already exists, no need to create it
		persistence.backingFilePath = configFilePath
		return nil
	}

	persistence.Lock()
	defer persistence.Unlock()

	persistence.backingFilePath = configFilePath
	if err = persistence.persistContextConfigInternal(defaultContextConfig); err != nil {
		return stacktrace.Propagate(err, "Failed to write default contexts config to file")
	}
	return nil
}

func (persistence *FileBackedConfigPersistence) persistContextConfigInternal(contextConfig *generated.KurtosisContextsConfig) error {
	serializedConfig, err := serializer.Marshal(contextConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to serialize content of contexts config object to JSON")
	}
	if err = os.WriteFile(persistence.backingFilePath, serializedConfig, defaultFilePerm); err != nil {
		return stacktrace.Propagate(err, "Unable to write context config file to '%s'",
			persistence.backingFilePath)
	}
	return nil
}

func getContextConfigFilePath() (string, error) {
	contextConfigFilePath, err := xdg.ConfigFile(path.Join(applicationDirname, fileName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to get context config file path")
	}
	return contextConfigFilePath, nil
}

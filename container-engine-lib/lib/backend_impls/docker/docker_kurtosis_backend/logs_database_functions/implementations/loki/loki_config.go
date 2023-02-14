package loki

import "fmt"

const (
	limitsRetentionPeriodHourIndicator = "h"
	tailMaxDurationHoursIndicator      = "h"
)

type LokiConfig struct {
	AuthEnabled   bool          `yaml:"auth_enabled"`
	Server        Server        `yaml:"server"`
	Common        Common        `yaml:"common"`
	StorageConfig StorageConfig `yaml:"storage_config"`
	SchemaConfig  SchemaConfig  `yaml:"schema_config"`
	Ruler         Ruler         `yaml:"ruler"`
	Compactor     Compactor     `yaml:"compactor"`
	LimitsConfig  LimitsConfig  `yaml:"limits_config"`
	Analytics     Analytics     `yaml:"analytics"`
	RuntimeConfig RuntimeConfig `yaml:"runtime_config"`
	Ingester      Ingester      `yaml:"ingester"`
	Querier       Querier       `yaml:"querier"`
}

type Server struct {
	HTTPListenPort uint16 `yaml:"http_listen_port"`
	LogLevel       string `yaml:"log_level"`
}

type Filesystem struct {
	ChunksDirectory string `yaml:"chunks_directory"`
	RulesDirectory  string `yaml:"rules_directory"`
}

type Storage struct {
	Filesystem Filesystem `yaml:"filesystem"`
}

type Kvstore struct {
	Store string `yaml:"store"`
}

type Ring struct {
	Kvstore Kvstore `yaml:"kvstore"`
}

type Common struct {
	PathPrefix        string  `yaml:"path_prefix"`
	Storage           Storage `yaml:"storage"`
	ReplicationFactor int     `yaml:"replication_factor"`
	Ring              Ring    `yaml:"ring"`
}

type StorageConfig struct {
	DisableBroadIndexQueries bool `yaml:"disable_broad_index_queries"`
}

type Index struct {
	Prefix string `yaml:"prefix"`
	Period string `yaml:"period"`
}

type Configs struct {
	From        string `yaml:"from"`
	Store       string `yaml:"store"`
	ObjectStore string `yaml:"object_store"`
	Schema      string `yaml:"schema"`
	Index       Index  `yaml:"index"`
}

type SchemaConfig struct {
	Configs []Configs `yaml:"configs"`
}

type Ruler struct {
	AlertmanagerURL string `yaml:"alertmanager_url"`
}

type Ingester struct {
	Wal IngesterWal `yaml:"wal"`
}

type IngesterWal struct {
	Enabled            bool   `yaml:"enabled"`
	Directory          string `yaml:"dir"`
	FlushOnShutdown    bool   `yaml:"flush_on_shutdown"`
	CheckpointDuration string `yaml:"checkpoint_duration"`
}

type Compactor struct {
	WorkingDirectory           string `yaml:"working_directory"`
	RetentionEnabled           bool   `yaml:"retention_enabled"`
	RetentionDeleteDelay       string `yaml:"retention_delete_delay"`
	RetentionDeleteWorkerCount int    `yaml:"retention_delete_worker_count"`
	DeletionMode               string `yaml:"deletion_mode"`
}

type LimitsConfig struct {
	RetentionPeriod string `yaml:"retention_period"`
	AllowDeletes    bool   `yaml:"allow_deletes"`
}

type Analytics struct {
	ReportingEnabled bool `yaml:"reporting_enabled"`
}

type RuntimeConfig struct {
	File   string `yaml:"file"`
	Period string `yaml:"period"`
}

type Querier struct {
	TailMaxDuration string `yaml:"tail_max_duration"`
}

// The following Loki configuration values are specific for the Kurtosis centralized logs Loki implementation
// some values were suggested by the Loki's documentation and this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2
func newDefaultLokiConfigForKurtosisCentralizedLogs(httpPortNumber uint16) *LokiConfig {
	newConfig := &LokiConfig{
		AuthEnabled: authEnabled,
		Server: Server{
			HTTPListenPort: httpPortNumber,
			LogLevel:       logLevel,
		},
		Common: Common{
			PathPrefix: dirpath,
			Storage: Storage{
				Filesystem: Filesystem{
					ChunksDirectory: chunksDirpath,
					RulesDirectory:  rulesDirpath,
				},
			},
			ReplicationFactor: replicationFactor,
			Ring: Ring{
				Kvstore: Kvstore{
					Store: ringKvStore,
				},
			},
		},
		StorageConfig: StorageConfig{
			DisableBroadIndexQueries: storageConfigShouldDisableBroadIndexQueries,
		},
		SchemaConfig: SchemaConfig{
			Configs: []Configs{
				{
					From:        schemaConfigFrom,
					Store:       schemaConfigStore,
					ObjectStore: schemaConfigObjectStore,
					Schema:      schemaConfigSchemaVersion,
					Index: Index{
						Prefix: schemaConfigIndexPrefix,
						Period: schemaConfigIndexPeriod,
					},
				},
			},
		},
		Ruler: Ruler{
			AlertmanagerURL: "",
		},
		Compactor: Compactor{
			WorkingDirectory:           compactorWorkingDirectory,
			RetentionEnabled:           compactorRetentionEnabled,
			RetentionDeleteDelay:       compactorRetentionDeleteDelay,
			RetentionDeleteWorkerCount: compactorRetentionDeleteWorkerCount,
			DeletionMode:               compactorDeletionMode,
		},
		LimitsConfig: LimitsConfig{
			RetentionPeriod: fmt.Sprintf("%v%v", LimitsRetentionPeriodHours, limitsRetentionPeriodHourIndicator),
			AllowDeletes:    allowDeletes,
		},
		Analytics: Analytics{
			ReportingEnabled: analyticsEnabled,
		},
		RuntimeConfig: RuntimeConfig{
			File:   runtimeConfigFilepath,
			Period: runtimeConfigPeriod,
		},
		Ingester: Ingester{
			Wal: IngesterWal{
				Enabled:            enableIngesterWal,
				Directory:          ingesterWalDirpath,
				FlushOnShutdown:    flushIngesterWalOnShutdown,
				CheckpointDuration: checkpointDuration,
			},
		},
		Querier: Querier{
			TailMaxDuration: fmt.Sprintf("%v%v", TailMaxDurationHours, tailMaxDurationHoursIndicator),
		},
	}
	return newConfig
}

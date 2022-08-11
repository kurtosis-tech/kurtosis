package loki

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
}

type Server struct {
	HTTPListenPort uint16 `yaml:"http_listen_port"`
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
	From        string   `yaml:"from"`
	Store       string `yaml:"store"`
	ObjectStore string `yaml:"object_store"`
	Schema string `yaml:"schema"`
	Index  Index  `yaml:"index"`
}

type SchemaConfig struct {
	Configs []Configs `yaml:"configs"`
}

type Ruler struct {
	AlertmanagerURL string `yaml:"alertmanager_url"`
}

type Compactor struct {
	WorkingDirectory           string `yaml:"working_directory"`
	RetentionEnabled           bool   `yaml:"retention_enabled"`
	RetentionDeleteDelay       string `yaml:"retention_delete_delay"`
	RetentionDeleteWorkerCount int    `yaml:"retention_delete_worker_count"`
}

type LimitsConfig struct {
	RetentionPeriod string `yaml:"retention_period"`
}

type Analytics struct {
	ReportingEnabled bool `yaml:"reporting_enabled"`
}

type RuntimeConfig struct {
	File   string `yaml:"file"`
	Period string `yaml:"period"`
}

func newDefaultLokiConfigForKurtosisCentralizedLogs() *LokiConfig {
	newConfig := &LokiConfig{
		AuthEnabled: authEnabled,
		Server: Server{
			HTTPListenPort: httpPortNumber,
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
			DisableBroadIndexQueries: storageConfigDisabledBroadIndexQueries,
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
		Compactor: Compactor{
			WorkingDirectory:           compactorWorkingDirectory,
			RetentionEnabled:           compactorRetentionEnabled,
			RetentionDeleteDelay:       compactorRetentionDeleteDelay,
			RetentionDeleteWorkerCount: compactorRetentionDeleteWorkerCount,
		},
		LimitsConfig: LimitsConfig{
			RetentionPeriod: limitsRetentionPeriod,
		},
		Analytics: Analytics{
			ReportingEnabled: analyticsEnabled,
		},

	}
	return newConfig
}

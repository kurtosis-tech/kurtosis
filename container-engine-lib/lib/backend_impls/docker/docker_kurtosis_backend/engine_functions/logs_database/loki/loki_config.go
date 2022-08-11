package loki

const (
	//The following configuration values are the default ones suggested by the
	//Loki's documentation and this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2

	//We enable multi-tenancy mode and we scope enclaves as tenants EnclaveId = TenantID (see more about multi-tenancy here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/)
	lokiDefaultAuthEnabled = true
	//The destinations path where the index, chnunks and rules will be saved
	LokiDefaultDirpath               = "/loki"
	lokiDefaultChunksDirectory = LokiDefaultDirpath + "/chunks"
	lokiDefaultRulesDirectory = LokiDefaultDirpath + "/rules"
	//We are going to run only one instance of Loki because we are using the filesystem storage, and it's ok for our current use case (https://grafana.com/docs/loki/latest/operations/storage/filesystem/#high-availability)
	lokiDefaultReplicationFactor = 1
	//it's the backend storage used for the ring
	lokiDefaultRingKvStore = "inmemory"
	//It's the default date when we started implementing Loki in Kurtosis
	lokiDefaultSchemaConfigFrom = "2022-08-01"
	//Boltdb-shipper is the adapter used to store the index in the chunk store (https://grafana.com/docs/loki/latest/fundamentals/architecture/#single-store)
	lokiDefaultSchemaConfigStore = "boltdb-shipper"
	//We are going to store the data in the container's filesystem (https://grafana.com/docs/loki/latest/operations/storage/filesystem)
	lokiDefaultSchemaConfigObjectStore = "filesystem"
	//The v12 generate lighter files compare with the previous versions (see video min 10:34)
	lokiDefaultSchemaConfigSchemaVersion = "v12"
	lokiDefaultSchemaConfigIndexPrefix = "index_"
	//24h is the max value allowed for boltdb-shipper
	lokiDefaultSchemaConfigIndexPeriod = "24h"
	//Enabled this feature because we are using boltdb-shipper index type, it helps to improve performance in expensive queries (see video min 40:00)
	lokiDefaultStorageConfigDisabledBroadIndexQueries = true
	//We don't want to send analytics metrics to Loki because it will be used pretty much for developing and testing purpose
	lokiDefaultAnalyticsEnabled = false

	//The next values are used to configure the retention period which is the method that we can use to delete old logs
	//We are going to store logs for 1 week = 168h. See more here: https://grafana.com/docs/loki/latest/operations/storage/retention/
	lokiDefaultCompactorWorkingDirectory = LokiDefaultDirpath + "/compactor"
	lokiDefaultCompactorSharedStore = "filesystem"
	lokiDefaultCompactorCompactionInterval = "10m"
	lokiDefaultCompactorRetentionEnabled = true
	lokiDefaultCompactorRetentionDeleteDelay = "2h"
	lokiDefaultCompactorRetentionDeleteWorkerCount = 150
	lokiDefaultLimitsRetentionPeriod = "168h"
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
	SharedStore                string `yaml:"shared_store"`
	CompactionInterval         string `yaml:"compaction_interval"`
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

func newDefaultLokiConfigForKurtosisCentralizedLogs(
	httpListenPort uint16,
) *LokiConfig {

	newConfig := &LokiConfig{
		AuthEnabled: lokiDefaultAuthEnabled,
		Server: Server{
			HTTPListenPort: httpListenPort,
		},
		Common: Common{
			PathPrefix: LokiDefaultDirpath,
			Storage: Storage{
				Filesystem: Filesystem{
					ChunksDirectory: lokiDefaultChunksDirectory,
					RulesDirectory:  lokiDefaultRulesDirectory,
				},
			},
			ReplicationFactor: lokiDefaultReplicationFactor,
			Ring: Ring{
				Kvstore: Kvstore{
					Store: lokiDefaultRingKvStore,
				},
			},
		},
		StorageConfig: StorageConfig{
			DisableBroadIndexQueries: lokiDefaultStorageConfigDisabledBroadIndexQueries,
		},
		SchemaConfig: SchemaConfig{
			Configs: []Configs{
				{
					From:        lokiDefaultSchemaConfigFrom,
					Store:       lokiDefaultSchemaConfigStore,
					ObjectStore: lokiDefaultSchemaConfigObjectStore,
					Schema:      lokiDefaultSchemaConfigSchemaVersion,
					Index: Index{
						Prefix: lokiDefaultSchemaConfigIndexPrefix,
						Period: lokiDefaultSchemaConfigIndexPeriod,
					},
				},
			},
		},
		Compactor: Compactor{
			WorkingDirectory:           lokiDefaultCompactorWorkingDirectory,
			SharedStore:                lokiDefaultCompactorSharedStore,
			CompactionInterval:         lokiDefaultCompactorCompactionInterval,
			RetentionEnabled:           lokiDefaultCompactorRetentionEnabled,
			RetentionDeleteDelay:       lokiDefaultCompactorRetentionDeleteDelay,
			RetentionDeleteWorkerCount: lokiDefaultCompactorRetentionDeleteWorkerCount,
		},
		LimitsConfig: LimitsConfig{
			RetentionPeriod: lokiDefaultLimitsRetentionPeriod,
		},
		Analytics: Analytics{
			ReportingEnabled: lokiDefaultAnalyticsEnabled,
		},
	}

	return newConfig
}

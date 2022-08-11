package loki

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"

const (
	////////////////////////--LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage           = "grafana/loki:main-19c7315"
	httpPortNumber uint16 = 3100 // Default Loki HTTP API port number, more here: https://grafana.com/docs/loki/latest/api/
	httpPortProtocol        = port_spec.PortProtocol_TCP

	configFilepath = "/etc/loki/local-config.yaml"
	binaryFilepath = "/usr/bin/loki"
	configFileFlag = "-config.file"
	////////////////////////--FINISH LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--LOKI CONFIGURATION SECTION--/////////////////////////////
	//The following Loki configuration values are specific for the Kurtosis centralized logs Loki implementation
	//some values were suggested by the Loki's documentation and this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2

	//We enable multi-tenancy mode, and we scope enclaves as tenants EnclaveId = TenantID (see more about multi-tenancy here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/)
	authEnabled = true
	//The destinations path where the index, chnunks and rules will be saved
	dirpath            = "/loki"
	chunksDirpath = dirpath + "/chunks"
	rulesDirpath  = dirpath + "/rules"
	//We are going to run only one instance of Loki because we are using the filesystem storage, it's ok for our current use case (https://grafana.com/docs/loki/latest/operations/storage/filesystem/#high-availability)
	replicationFactor = 1
	//it's the backend storage used for the ring
	ringKvStore = "inmemory"
	//It's the date when we started implementing Loki in Kurtosis
	schemaConfigFrom = "2022-08-01"
	//Boltdb-shipper is the adapter used to store the index in the chunk store (https://grafana.com/docs/loki/latest/fundamentals/architecture/#single-store)
	//we use boltdb-shipper because it is the only one that allows us to set the Compactor's retention logs byt Tenant (https://grafana.com/docs/loki/latest/operations/storage/retention/)
	schemaConfigStore = "boltdb-shipper"
	//We are going to store the data in the container's filesystem (https://grafana.com/docs/loki/latest/operations/storage/filesystem)
	schemaConfigObjectStore = "filesystem"
	//The v12 generate lighter files compare with the previous versions (see video min 10:34)
	schemaConfigSchemaVersion = "v12"
	schemaConfigIndexPrefix   = "index_"
	//24h is the max value allowed for boltdb-shipper
	schemaConfigIndexPeriod = "24h"
	//Enabled this feature because we are using boltdb-shipper index type, it helps to improve performance in expensive queries (see video min 40:00)
	storageConfigDisabledBroadIndexQueries = true
	//We don't want to send analytics metrics to Loki because it will be used pretty much for developing and testing purpose
	analyticsEnabled = false

	//The next values are used to configure the Compactor component which allows us to enable the retention period
	//See more here: https://grafana.com/docs/loki/latest/operations/storage/retention/
	compactorWorkingDirectory                = dirpath + "/compactor"
	compactorRetentionEnabled                      = true
	compactorRetentionDeleteDelay       = "2h"
	compactorRetentionDeleteWorkerCount = 150
	//It's the global retention period then we will set retention periods by TenantID that overrides this value
	//the global retention period store logs for 1 week = 168h.
	limitsRetentionPeriod = "168h"
	////////////////////////--FINISH--LOKI CONFIGURATION SECTION--/////////////////////////////
)

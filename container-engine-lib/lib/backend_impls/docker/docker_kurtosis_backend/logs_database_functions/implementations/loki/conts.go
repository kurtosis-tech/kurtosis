package loki

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	configDirpath = "/etc/loki/"

	////////////////////////--LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage          = "grafana/loki:main-19c7315"
	httpPortNumber   uint16 = 3100 // Default Loki HTTP API port number, more here: https://grafana.com/docs/loki/latest/api/
	httpPortProtocol        = port_spec.PortProtocol_TCP

	configFilepath = configDirpath + "local-config.yaml"
	binaryFilepath = "/usr/bin/loki"
	configFileFlag = "-config.file"
	////////////////////////--FINISH LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--LOKI CONFIGURATION SECTION--/////////////////////////////
	//The following Loki configuration values are specific for the Kurtosis centralized logs Loki implementation
	//some values were suggested by the Loki's documentation and this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2

	//We enable multi-tenancy mode, and we scope enclaves as tenants EnclaveId = TenantID (see more about multi-tenancy here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/)
	authEnabled = true
	//The destinations path where the index, chunks and rules will be saved
	dirpath       = "/loki"
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
	//another benefit of using boltdb-shipper is that the Log entry deletion is supported only when the BoltDB Shipper is configured for the index store. (https://grafana.com/docs/loki/latest/operations/storage/logs-deletion/#log-entry-deletion)
	schemaConfigStore = "boltdb-shipper"
	//We are going to store the data in the container's filesystem (https://grafana.com/docs/loki/latest/operations/storage/filesystem)
	schemaConfigObjectStore = "filesystem"
	//The v12 generate lighter files compare with the previous versions (see min 40:00 of this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2)
	schemaConfigSchemaVersion = "v12"
	schemaConfigIndexPrefix   = "index_"
	//24h is the max value allowed for boltdb-shipper
	//This value is also need because retention is only available if the index period is 24h. (see more here: https://grafana.com/docs/loki/latest/operations/storage/retention/#compactor)
	schemaConfigIndexPeriod = "24h"
	//Enabled this feature because we are using boltdb-shipper index type, it helps to improve performance in expensive queries (see min 10:34 of this video: https://grafana.com/go/webinar/logging-with-loki-essential-configuration-settings/?pg=docs-loki&plcmt=footer-resources-2))
	storageConfigShouldDisableBroadIndexQueries = true
	//We don't want to send analytics metrics to Loki because it will be used pretty much for developing and testing purpose
	analyticsEnabled = false

	//The next values are used to configure the Compactor component which allows us to enable the retention period
	//See more here: https://grafana.com/docs/loki/latest/operations/storage/retention/
	compactorWorkingDirectory           = dirpath + "/compactor"
	compactorRetentionEnabled           = true
	//More about retention delete delay here: https://grafana.com/docs/loki/latest/operations/storage/retention/#compactor
	compactorRetentionDeleteDelay       = "1h"
	compactorRetentionDeleteWorkerCount = 150
	//The "filter-and-delete" mode remove the logs from the storage (see more here: https://grafana.com/docs/loki/latest/configuration/#compactor and here: https://grafana.com/docs/loki/latest/operations/storage/logs-deletion/#configuration)
	compactorDeletionMode = "filter-and-delete"
	//It's the global retention period (the retention period by TenantID overrides this value)
	//the global retention period store logs for 30 days = 720h.
	limitsRetentionPeriod = "720h"
	//This value enables the deletion API
	allowDeletes = true

	//The filepath of the runtime configuration that we are going to use for limits retention period by TenantID
	//see more here: https://grafana.com/docs/loki/latest/configuration/#runtime-configuration-file
	runtimeConfigFilepath = configDirpath + "runtime-config.yaml"
	//How often to check the file.
	runtimeConfigPeriod             = "20s"
	runtimeConfigFileInitialContent = "overrides:"

	//The configuration for the ingester WAL, it's important for storing chunks when the server is shutdown. See more here: https://grafana.com/docs/loki/latest/configuration/#ingester
	enableIngesterWal = true
	ingesterWalDirpath  = dirpath + "/wal"
	flushIngesterWalOnShutdown = true //It's useful for graceful shutdowns
	checkpointDuration = "1s" //It's useful for ungraceful shutdowns, whe the server is restarted the WAL loads the last checkpoint saved
	////////////////////////--FINISH--LOKI CONFIGURATION SECTION--/////////////////////////////
)

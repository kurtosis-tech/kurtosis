package traefik

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
)

const (
	////////////////////////--TRAEFIK CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "traefik:2.10.6"

	configDirpath  = "/etc/traefik/"
	configFilepath = configDirpath + "traefik.yml"
	binaryFilepath = "/usr/local/bin/traefik"
	////////////////////////--FINISH TRAEFIK CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--TRAEFIK CONFIGURATION SECTION--/////////////////////////////
	traefikNetworkid = consts.NameOfNetworkToStartEngineAndLogServiceContainersIn

	configFileTemplateName = "traefikConfigFileTemplate"

	configFileTemplate = `
api:
  dashboard: true
  insecure: true
  disabledashboardad: true
  
entryPoints:
  web:
    address: ":{{ .WebAddress }}"
  traefik:
    address: ":{{ .TraefikAddress }}"
  
providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: "{{ .NetworkId }}"
`
	////////////////////////--FINISH--TRAEFIK CONFIGURATION SECTION--/////////////////////////////
)

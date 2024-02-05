package traefik

const (
	////////////////////////--TRAEFIK CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "traefik:2.10.6"

	configDirpath  = "/etc/traefik/"
	configFilepath = configDirpath + "traefik.yml"
	binaryFilepath = "/usr/local/bin/traefik"
	////////////////////////--FINISH TRAEFIK CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--TRAEFIK CONFIGURATION SECTION--/////////////////////////////
	configFileTemplate = `
accesslog: {}
log:
  level: DEBUG
api:
  debug: true
  dashboard: true
  insecure: true
  disabledashboardad: true
 
entryPoints:
  web:
    address: ":{{ .HttpPort }}"
  traefik:
    address: ":{{ .DashboardPort }}"

providers:
  docker:
    endpoint: "unix:///var/run/podman/podman.sock"
    exposedByDefault: false
    network: "{{ .NetworkId }}"
`
	////////////////////////--FINISH--TRAEFIK CONFIGURATION SECTION--/////////////////////////////
)

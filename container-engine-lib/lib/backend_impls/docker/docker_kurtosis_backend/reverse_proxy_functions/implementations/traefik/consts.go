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
  # api over the traefik endpoint
  insecure: true
  disabledashboardad: true
 
entryPoints:
  # http traffic
  web:
    address: ":{{ .HttpPort }}"
  # API endpoint
  traefik:
    address: ":{{ .DashboardPort }}"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    # we don't want the containers to be exposed by default.
    # we are enabling Traefik at the container level instead.
    exposedByDefault: false
    # Docker network to start Traefik in.
    network: "{{ .NetworkId }}"
`
	////////////////////////--FINISH--TRAEFIK CONFIGURATION SECTION--/////////////////////////////
)

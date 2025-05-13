package v5

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

type GrafanaLokiConfigV5 struct {
	// ShouldStartBeforeEngine starts Grafana and Loki before the engine, if true.
	// Equivalent to running `grafloki start` before `engine start`.
	// Useful for treating Grafana and Loki as default logging setup in Kurtosis.
	ShouldStartBeforeEngine bool   `yaml:"should-start-before-engine,omitempty"`
	GrafanaImage            string `yaml:"grafana-image,omitempty"`
	LokiImage               string `yaml:"loki-image,omitempty"`
}

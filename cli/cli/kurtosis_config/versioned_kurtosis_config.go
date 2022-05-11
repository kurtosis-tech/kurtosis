package kurtosis_config

type VersionedKurtosisConfig interface {
	Validate() error
}

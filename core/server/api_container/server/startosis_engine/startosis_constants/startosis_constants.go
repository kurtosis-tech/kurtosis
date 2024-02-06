package startosis_constants

type StarlarkContextParam string

const (
	MainFileName     = "main.star"
	KurtosisYamlName = "kurtosis.yml"
	EmptyInputArgs   = "{}" // empty JSON

	NoOutputObject = ""

	PackageIdPlaceholderForStandaloneScript                          = "DEFAULT_PACKAGE_ID_FOR_SCRIPT"
	PlaceHolderMainFileForPlaceStandAloneScript                      = ""
	ParallelismParam                            StarlarkContextParam = "PARALLELISM"

	// DefaultPersistentDirectorySize 1Gi Megabytes is the default value and what most drivers support
	DefaultPersistentDirectorySize int64 = 1024 * 1024 * 1024
)

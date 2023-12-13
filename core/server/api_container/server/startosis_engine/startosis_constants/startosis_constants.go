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

	// DefaultPersistentDirectorySize 500 Megabytes is the default value
	DefaultPersistentDirectorySize int64 = 500 * 1024 * 1024
)

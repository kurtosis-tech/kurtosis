package repl_consts

type ReplType string
const (
	KurtosisSocketEnvVar          = "KURTOSIS_API_SOCKET"
	EnclaveDataMountDirpathEnvVar = "ENCLAVE_DATA_DIR_MOUNTPOINT"

	ReplType_Javascript ReplType = "javascript"

	javascriptPackageInstallationDirpath = "/preinstalled-node-modules"
	javascriptInstalledPackagesDirpath   = javascriptPackageInstallationDirpath + "/node_modules"
)
// Mapping of repl_type -> dirpath_where_package_installation_cmd_should_be_run
var PackageInstallationDirpaths = map[ReplType]string{
	ReplType_Javascript: javascriptPackageInstallationDirpath,
}

// Mapping of repl_type -> dirpath_where_packages_actually_are_installed
var InstalledPackagesDirpath = map[ReplType]string{
	ReplType_Javascript: javascriptInstalledPackagesDirpath,
}

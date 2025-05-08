package snapshots

import "path"

const (
	SnapshotDir = "/kurtosis-data/snapshot-store"
)

var (
	SnapshotDirPath                      = path.Join(SnapshotDir, "snapshot")
	SnapshotTmpDirPath                   = path.Join(SnapshotDirPath, "tmp")
	SnapshotServicesDirPath              = path.Join(SnapshotDirPath, "services")
	SnapshotFilesArtifactsDirPath        = path.Join(SnapshotDirPath, "files-artifacts")
	SnapshotPersistentDirectoriesDirPath = path.Join(SnapshotDirPath, "persistent-directories")

	serviceStartupOrderFileName = "service-startup-order.txt"

	filesArtifactsNamesFileName = "files-artifacts-names.txt"
	filesArtifactsNamesFileDir  = "files-artifacts"

	persistentDirectoriesDirName = "persistent-directories"

	serviceConfigFileName         = "service-config.json"
	serviceConfigPathFmtSpecifier = "services/%s/%s"
)

package snapshots

import "os"

// Snapshot package layout
// /snapshot
//
//	/persistent-directories
//	   /persistent-key-1
//		   /tar.tgz
//	   /persistent-key-2
//		   /tar.tgz
//	/files-artifacts
//		files-artifacts-name.tar.tgz
//		...
//	/services
//		/service-name
//			service-config.json
//			image.tar
//			service-registration.json
//	 args.json
//	 return args ...
//	 service-startup-order.txt
//	 files-artifacts-names.txt
//	 persistent-directories-names.txt
const (
	snapshotServicesDirPath              = "services"
	snapshotFilesArtifactsDirPath        = "files-artifacts"
	snapshotPersistentDirectoriesDirPath = "persistent-directories"

	serviceStartupOrderFileName = "service-names.txt"

	snapshotDirPerms = os.FileMode(0777) // TODO: refactor to use enclave data directory perms

	// filesArtifactsNamesFileName = "files-artifacts-names.txt"
	// filesArtifactsNamesFileDir  = "files-artifacts"

	// persistentDirectoriesDirName = "persistent-directories"

	serviceConfigFileName         = "service-config.json"
	serviceConfigPathFmtSpecifier = "services/%s/%s"
	serviceImagePathFmtSpecifier  = "services/%s/%s"
)

var (
	snapshottedImageNameFmtSpecifier = "%s-snapshotted-img.tar"
)

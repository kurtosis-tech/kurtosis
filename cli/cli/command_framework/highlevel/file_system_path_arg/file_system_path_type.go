package file_system_path_arg

//go:generate go run github.com/dmarkham/enumer -type=FileSystemPathType -trimprefix=FileSystemPathType_ -transform=snake
type FileSystemPathType int
const (
	FileSystemPathType_Filepath FileSystemPathType = iota
	FileSystemPathType_Dirpath
	FileSystemPathType_FilepathOrDirpath
)


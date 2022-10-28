package startosis_modules

// ModuleContentProvider A module content provider allows you to get a Startosis module given a url
// It fetches the contents of the module for you
type ModuleContentProvider interface {
	GetOnDiskAbsoluteFilePath(string) (string, error)

	GetModuleContents(string) (string, error)
	StoreModuleContents(string, []byte, bool) (string, error)
}

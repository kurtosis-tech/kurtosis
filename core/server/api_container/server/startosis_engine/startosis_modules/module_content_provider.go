package startosis_modules

// ModuleContentProvider A module content provider allows you to get a Startosis module given a url
// It fetches the contents of the module for you
type ModuleContentProvider interface {
	GetModuleContents(string) (string, error)
	GetFileAtRelativePath(fileBeingInterpreted string, relFilepathOfFileToRead string) (string, error)
	IsGithubPath(path string) bool
}

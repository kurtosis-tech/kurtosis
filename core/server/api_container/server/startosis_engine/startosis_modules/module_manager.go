package startosis_modules

// ModuleManager A module manager allows you to get a Startosis module given a url
// It fetches the contents of the module for you
type ModuleManager interface {
	GetModule(string) (string, error)
}

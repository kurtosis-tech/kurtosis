package startosis_modules

// A Kurtosis command is a command that wraps a Cobra command to make it easier to work with
// There are many implementations, and some wrap the logic of others
type ModuleManager interface {
	GetModule(string) (string, error)
}

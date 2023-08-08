package startosis_validator

//go:generate go run github.com/dmarkham/enumer -trimprefix "ComponentExistence" -type=ComponentExistence
type ComponentExistence uint8

const (
	// ComponentNotFound - Component does not exist in the enclave
	ComponentNotFound ComponentExistence = iota

	// ComponentExistedBeforePackageRun - The component existed in the enclave at the beginning of the package run.
	ComponentExistedBeforePackageRun

	// ComponentCreatedOrUpdatedDuringPackageRun - The component was created or updated during this run.
	ComponentCreatedOrUpdatedDuringPackageRun
)

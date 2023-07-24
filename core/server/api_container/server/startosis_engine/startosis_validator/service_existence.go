package startosis_validator

//go:generate go run github.com/dmarkham/enumer -trimprefix "ServiceExistence" -type=ServiceExistence
type ServiceExistence uint8

const (
	// ServiceNotFound - Service does not exist in the enclave
	ServiceNotFound ServiceExistence = iota

	// ServiceExistedBeforePackageRun - The service existed in the enclave at the beginning of the package run.
	ServiceExistedBeforePackageRun

	// ServiceCreatedOrUpdatedDUringPackageRun - The service was created or updated during this run.
	ServiceCreatedOrUpdatedDuringPackageRun
)

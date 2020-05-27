package testsuite

type TestSuite interface {
	// Intended to be implemented by the user to register whatever tests they please
	GetTests() map[string]TestConfig
}

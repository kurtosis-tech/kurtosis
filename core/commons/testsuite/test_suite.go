package testsuite

/*
An interface which the user implements to register their tests.
 */
type TestSuite interface {
	// Get all the tests in the test suite; this is where users will "register" their tests
	GetTests() map[string]Test
}

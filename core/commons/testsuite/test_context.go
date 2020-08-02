package testsuite

/*
An object that will be passed in to every test, which the user can use to manipulate the results of the test
 */
type TestContext struct {}

/*
Fails the test with the given error
 */
func (context TestContext) Fatal(err error) {
	// We rely on panicking here because we want to completely stop whatever the test is doing
	failTest(err)
}

/*
Asserts that the given condition is true, and if not then fails the test and returns the given error
 */
func (context TestContext) AssertTrue(condition bool, err error) {
	if (!condition) {
		failTest(err)
	}
}

func failTest(err error) {
	panic(err)
}

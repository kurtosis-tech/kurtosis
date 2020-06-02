package testsuite

import "fmt"

// An object that will be passed in to every test, which the user can use to manipulate the results of the test
// NOTE: This object's methods rely on pankcing on error because we want to completely abort whatever the test is doing
type TestContext struct {}

// TODO either: 1) a lot more useful methods here or 2) a way to leverage Go's inbuilt testing framework to do this testing

func (context TestContext) Fatal(err error) {
	// We rely on panicking here because we want to completely stop whatever the test is doing
	panic(err)
}

func (context TestContext) AssertTrue(condition bool) {
	if (!condition) {
		panic(fmt.Sprintf("Assertion failed"))
	}
}

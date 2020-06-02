package testsuite

import "fmt"

// An object that will be passed in to every test, which the user can use to manipulate the results of the test
type TestContext struct {}

func (context TestContext) Fatal(err error) {
	panic(err)
}

func (context TestContext) AssertTrue(condition bool) {
	if (!condition) {
		panic(fmt.Sprintf("Assertion failed"))
	}
}

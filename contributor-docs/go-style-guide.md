Go Style Guide
==============
Contains best practices for writing safe Go code.

### All logging should go through Logrus
* Rationale: gives us a single point of controlling output (just change the `logrus.StandardLogger().Out`)
* For regular log messages, use `logrus.Infof`, `logrus.Debugf`, etc.
* For printing raw text, use `fmt.Fprintln(logrus.StandardLogger().Out, someThingIWantToPrint)`

### Use `stacktrace.Propagate` or `stacktrace.NewError` from the `github.com/kurtosis-tech/stacktrace` package whenever possible
Doing so makes error handling and formatting mindless, and stacktraces are auto-generated so debugging is easy. Examples of what this should look like:
```golang
// Case 1: the function we're calling on returns an error
if err := funcThatOnlyReturnsErr(); err != nil {
    return stacktrace.Propagate(err, "An error occurred when ...describe the thing...")
}

// Case 2: the function returns an error and a value
value, err := funcThatReturnsValAndErr()
if err != nil {
    return stacktrace.Propagate(err, "An error occurred when ...describe the thing...")
}

// Case 3: we want to throw an error but aren't propagating an error
if len(someMap) == 0 {
    return stacktrace.NewError("Expected the map to have at least one element, but it had none")
}

// Case 4: function takes parameters
if err := someFuncWithParameters(value1, value2); err != nil {
    return stacktrace.Propagate(err, "An error occurred calling someFuncWithParameters with param1 '%v' and param2 '%v'", value1, value2)
}
```

The ONLY 1% of the time you shouldn't use this is if you don't actually want the stacktrace.

### Use the defer-undo pattern to implement resource cleanup
See the page on [picking up your toys](./style-guide/picking-up-your-toys.md) on how to use Go's `defer` statement to implement correct resource ownership in Go.

### Use the enumer tool for enums
Go sadly doesn't support enums out of the box. It allows you to sorta kinda emulate them, but they lack basic functionality like "string -> enum" and "enum -> string".

Fortunately, [the "enumer" tool](https://github.com/dmarkham/enumer) was built for exactly this purpose. Thus, when creating an enum, use the enumer tool.

For example:

```golang
//go:generate go run github.com/dmarkham/enumer -type=Weekday -transform=snake-upper
type Weekday int
const (
    Monday Weekday = iota
    Tuesday
    Wednesday
    Thursday
    Friday
    Saturday
    Sunday
)
```

NOTE: To get Go to properly track the dependency on the enumer package, you'll need to add the following file (which is a Go convention) to the root of your repo:

```golang
// +build tools

package main

// This file is to make sure that go.mod continues to record the tools we need in our 'go generate' commands
// See: https://marcofranssen.nl/manage-go-tools-via-go-modules
import (
    _ "github.com/dmarkham/enumer"
)
```

### Use the `var` section for constant maps, lists, and structs
Go sadly doesn't allow maps, lists, or structs in the `const` section. We use the `var` section to emulate this. **Unfortunately, variables declared in `var` are modifiable, so we rely on Kurtosian devs to treat those maps, lists and structs declared in `var` as constants and not modify them!**

```go
var (
    dockerContainerStatusToKurtosisContainerStatus = map[docker.ContainerStatus]kurtosis.ContainerStatus{
        docker.Running: kurtosis.ContainerStatus_Running,
        docker.Stopped: kurtosis.ContainerStatus_Stopped,
        docker.Paused: kurtosis.ContainerStatus_Paused,
    }
)
```

### Do not use Go's named return values!
Go provides the ability to name return values of a function. For example:

```go
func DoSomething() (string, string, string, error)
```

becomes:

```go
func DoSomething() (containerId string, podId string, serviceId string, err error)
```

HOWEVER, if you ever assign something to those return values, e.g.:

```go
containerId := getContainerId()
```

Then you've actually assigned to the return value of the function! This can cause all kinds of subtle bugs, so we ban the use of named return values.

NOTE: We used to use named return values prior to October, 2022 for documentation purposes by prefixing `result` to reduce the collision likelihood; these should be removed anywhere that you see them.

### Handle errors, even if you're not required to
This forces the coder to think about and handle all cases, which writes safer and clearer code, which protects the Future Midnight Kurtosian.

### Always check the `ok` value of casts, and the `found` values 
Same idea - writing defensive code to protect the Future Midnight Kurtosian.

Checking casts:
```go
castedValue, ok := uncastedValue.(*MyStruct)
if !ok {
    return stacktrace.NewError("An error occurred casting to MyStruct")
}
```

Checking maps:
```go
someValue, found := myMap[someKey]
if !found {
    return stacktrace.NewError("Couldn't find key '%v' in the map", someKey)
}
```

### Be VERY careful with references to for loop iteration variables (ESPECIALLY with lambdas)!!
Go in-place modifies iteration variables of `for` loops, which when mixed with lambdas can cause [the subtle bug described here](https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/#:~:text=has%20unexpected%20effects.-,Closures,-Finally%20we%20come). This _seems_ rare, until you consider that `defer` often operates on a lambda. The solution, as described in the blog post, is to either make your lambda take in an argument, or (better yet) call a named helper function.

### Strongly prefer making struct fields private, and putting a constructor next to the struct
Making struct fields private follows the general principles of "make things as private as possible", and the idea of the constructor is so that adding/removing fields will cause a compile break in all places that use the constructor. This shortens feedback loops by forcing the Kurtosis dev who has the most context about added/removed field to handle the break.

### Always use by-reference (pointer) function receivers 
Go allows you to have either by-value or by-reference function receivers:
```go
// By value
func (myStruct MyStruct) DoSomething() {
    myStruct.property = "something"
}

// By reference (note the *MyStruct)
func (myStruct *MyStruct) DoSomething() {
    myStruct.property = "something"
}
```

While the by-value idea in Go is nice, in practice it opens the door bugs: the by-value example above will not actually change the `property` on `MyStruct` because `myStruct` is actually a copy. 

To obviate these bugs and reduce Kurtosian mental load, we always use by-reference pointer receivers. This is no different from languages like Java, where object references are _always_ by value.

### Prefer passing around pointers to structs rather than the structs themselves
For example:

```go
func MyFunc(myObj *TheObject)
```

rather than:

```go
func MyFunc(myObj TheObject)  // Note missing asterisk
```

Rationale:

1. When passing the struct itself, Go will copy the object by-value. As objects get bigger, this can end up being a lot of bytes copied on every function call.
2. It makes coding more brainless, as you don't need to think "should I return the struct, or a pointer to the struct?"


### Use type aliases for documentation
We don't use named return values (see above), but we _can_ use type aliases for very neat documentation. E.g.

```go
func DoSomething() (string, string, string, error) {
    // ...
}
```

is much cleaner as:

```go
type PodID string
type ContainerID string
type ServiceID string

func DoSomething() (PodID, ContainerID, ServiceID, error) {
    // ...
}
```

### Use `require.XXXXX` rather than `assert.XXXXX` when testing
This is a gotcha with the testing library we use, `testify` - `assert` will actually let the test keep running even if an assertion fails, while `require` will fail the test immediately.

In essence, use `require.True`, `require.NonNil`, etc. rather than `assert.True`, `assert.NonNil`, etc.

### To mock a struct, create an interface with a `MockXXXXX` implementation
Go doesn't neatly support creating mocks. To mock a struct, simply extract its signature into an interface (which your struct will implement due to Go's implicit interface implementation), create a `MockXXXXX` version of it, and use that.

### Prefer one file per struct ("file early, file often")
While Go allows multiple structs per file, one file per struct generally makes it easier to find things. This isn't a hard rule - oftentimes it makes sense to have little package structs in the same file, so use your judgement.

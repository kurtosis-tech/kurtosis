Picking up your toys: the rules of resource ownership
author @mieubrisse
=====================================================
Code often needs external resources - e.g. file handles or Docker containers. Because they are external, these resources have lifetimes independent of the program. In essence, these resources will leak if the program doesn't handle them on exit. We prevent resource leaks at Kurtosis with three rules of resource ownership:

1. A function owns the resources it creates
1. A function must clean up the resources it owns OR transfer ownership to the caller
1. A function may only transfer resource ownership upon success (not failure)

Example
-------
Take a look at the following example:

```go
func StartContainer(imageName string) error {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    return nil
}
```

This function seems fine, but it has two resource leaks. 

**Leak 1:** a successful call to `StartContainer` will create a container via `docker.CreateContainer`. When `StartContainer` ends, the container ID drops off the stack. The container ID is the only reference to the container, so the container is leaked. `StartContainer` does not pick up its toys: it owns the container at time of exit, but does not clean it up.

We fix this leak by transferring ownership to the caller at the end of the function:

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    return containerID, nil
}
```

The caller will receive ownership, and becomes responsible for cleaning up the container.

**Leak 2:** an error in `docker.WaitForContainerAvailability` causes `StartContainer` to exit without transferring ownership of the container. The container ID will drop off the stack, and the container will be leaked.

We fix this problem with the conditional `defer-undo` pattern after creating the container:

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }
    shouldDeleteContainer := true
    defer func() {
        if shouldDeleteContainer {
            docker.DeleteContainer(containerID)
        }
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    shouldDeleteContainer = false
    return containerID, nil
}
```

This pattern ensures that `StartContainer` cleans up the container if any error occurs. `StartContainer` now picks up its toys.

Deferred Cleanup Timing
-----------------------
### Deferred cleanup must come after error-checking
Notice how we scheduled the `defer-undo` _after_ the error-checking for `docker.CreateContainer`. The ordering "after" is important because of the third ownership rule: resource ownership is only transferred upon success. 

`docker.CreateContainer` obeys this rule; it will not return a container ID if it fails. `StartContainer` cannot `defer-undo` before error-checking or it will try to delete an empty container ID.

This is wrong:

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    // ERROR-CHECKING HASN'T BEEN PERFORMED, SO WE DON'T KNOW IF CONTAINER ID IS VALID
    shouldDeleteContainer := true
    defer func() {
        if shouldDeleteContainer {
            docker.DeleteContainer(containerID)
        }
    }
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    shouldDeleteContainer = false
    return containerID, nil
}
```

### Deferred cleanup must come IMMEDIATELY after error-checking
After we error-check `docker.CreateContainer`, `StartContainer` assumes ownership of the container. We cannot wait to `defer`, else we risk leaking the resource.

For example, this code will leak the container if `docker.WaitForContainerAvailability` errors:

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    shouldDeleteContainer := true
    defer func() {
        if shouldDeleteContainer {
            docker.DeleteContainer(containerID)
        }
    }

    shouldDeleteContainer = false
    return containerID, nil
}
```

We go so far as, "there should be no whitespace between the error-checking and the `defer`":

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    // THIS WHITESPACE IS WRONG

    shouldDeleteContainer := true
    defer func() {
        if shouldDeleteContainer {
            docker.DeleteContainer(containerID)
        }
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    shouldDeleteContainer = false
    return containerID, nil
}
```

We're strict like this because a Kurtosis dev with less context _will_ make changes in the future. No whitespace means nowhere to add code means no place to create an accidental bug.

### Deferred cleanup must always happen
It can be tempting to skip the `defer-undo` when no actions happen after resource creation. Doing so invites bugs. A Kurtosis dev with less context _will_ add new code, 

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }

    // ANY 'return' STATEMENTS HERE WILL LEAK THE CONTAINER

    return containerID, nil
}
```

Failed cleanups and user ownership
----------------------------------
Errors are often possible inside deferred code. We cannot guarantee successful cleanup, so Kurtosis deferred functions are best-effort by necessity. When a cleanup failure occurs, ownership of the resource falls to the user. Thus, when a cleanup failure occurs, we print a loud `ACTION REQUIRED` error log message with cleanup actions that the user must take.

This is what `StartContainer` would look like if `docker.DeleteContainer` can throw an error:

```go
func StartContainer(imageName string) (string, error) {
    containerID, err := docker.CreateContainer(imageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting container with image '%v'", imageName)
    }
    shouldDeleteContainer := true
    defer func() {
        if shouldDeleteContainer {
            if err := docker.DeleteContainer(containerID); err != nil {
                logrus.Errorf("We tried to clean up the container we started with image '%v' and ID '%v', but an error occurred during deletion:\n%v", imageName, containerID, err)
                logrus.Errorf("!!! ACTION REQUIRED !!!! You will need to delete container with ID '%v' manually!!!!", containerID)
            }
        }
    }

    if err := docker.WaitForContainerAvailability(containerID); err != nil {
        return "", stacktrace.Propagate(err, "Container '%v' didn't become available", containerID)
    }

    shouldDeleteContainer = false
    return containerID, nil
}
```

Notice how:

1. We print information about the container that failed (image and ID), so the user can find the resource
1. We print the `docker.DeleteContainer` error, so the user knows exactly why the failure occurred
1. We give the user the remediation steps to finish the cleanup

This pattern is standardized, and should be followed for any deferred-cleanup code that can return an error.

Repeatability
-------------
This pattern is repeatable, and harmonizes with the principle of "eject early, eject often". This is the same example, extended with a second container resource:

```go
func StartContainer(firstContainerImageName string, secondContainerImageName string) (string, string, error) {
    // Start first container
    firstContainerID, err := docker.CreateContainer(firstContainerImageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting the first container with image '%v'", firstContainerImageName)
    }
    shouldDeleteFirstContainer := true
    defer func() {
        if shouldDeleteFirstContainer {
            docker.DeleteContainer(firstContainerID)
        }
    }

    if err := docker.WaitForContainerAvailability(firstContainerID); err != nil {
        return "", stacktrace.Propagate(err, "First container '%v' didn't become available", firstContainerID)
    }

    // Start second container
    secondContainerID, err := docker.CreateContainer(secondContainerImageName)
    if err != nil {
        return "", stacktrace.Propagate(err, "An error occurred starting the second container with image '%v'", secondContainerImageName)
    }
    shouldDeleteSecondContainer := true
    defer func() {
        if shouldDeleteSecondContainer {
            docker.DeleteContainer(secondContainerID)
        }
    }

    if err := docker.WaitForContainerAvailability(secondContainerID); err != nil {
        return "", stacktrace.Propagate(err, "Second container '%v' didn't become available", secondContainerID)
    }

    // More resources could be added here

    shouldDeleteFirstContainer = false
    shouldDeleteSecondContainer = false
    return firstContainerID, secondContainerID, nil
}
```

Notes
-----
### In other languages
We used Go's `defer` in this example, but we can accomplish the same result in other languages with `try-finally`.

For example, Typescript:

```typescript
function startContainer(firstContainerImageName: string, secondContainerImageName: string): Result<[string, string], Error> {
    // Start first container
    const createFirstContainerResult: Result<string, Error> = docker.createContainer(firstContainerImageName)
    if (createFirstContainerResult.isErr()) {
        return err(createFirstContainerResult.error)
    }
    const firstContainerID: string = createFirstContainerResult.value
    let shouldDeleteFirstContainer: boolean = true
    try {
        const waitForFirstContainerAvailabilityResult: Result<null, Error> = docker.WaitForContainerAvailability(firstContainerID)
        if (waitForFirstContainerAvailabilityResult.isErr()) {
            return err(waitForFirstContainerAvailabilityResult.error)
        }

        // Start second container
        const createSecondContainerResult: Result<string, Error> = docker.createContainer(secondContainerImageName)
        if (createSecondContainerResult.isErr()) {
            return err(createSecondContainerResult.error)
        }
        const secondContainerID: string = createSecondContainerResult.value
        let shouldDeleteSecondContainer: boolean = true
        try {
            const waitForSecondContainerAvailabilityResult: Result<null, Error> = docker.WaitForContainerAvailability(secondContainerID)
            if (waitForSecondContainerAvailabilityResult.isErr()) {
                return err(waitForSecondContainerAvailabilityResult.error)
            }

            // NOTE HOW CONDITIONALS ARE SET TO FALSE HERE
            let shouldDeleteFirstContainer = true
            let shouldDeleteSecondContainer = true
            return ok([secondContainerID, secondContainerID])
        } finally {
            if shouldDeleteSecondContainer {
                docker.deleteContainer(secondContainerID)
            }
        }
    } finally {
        if shouldDeleteFirstContainer {
            docker.deleteContainer(firstContainerID)
        }
    }

    // NOTE HOW THE RETURN STATEMENT HAS BEEN MOVED INSIDE
}
```

The verbosity and extreme nesting is a good example of why Go's `defer` is nice.

### Rust
Rust programmers will notice similarities to how Rust handles ownership. This is not accidental; Rust does ownership very well.

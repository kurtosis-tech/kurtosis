Debugging Failed Tests
======================
Tests will of course fail over the course of your development, so here are some common error scenarios that you might encounter:

Tests failed but no controller logs were printed
------------------------------------------------
If your tests are failing but you're not getting any controller logs whatsoever, you either have a failure in launching the test's controller container or in the machinery for reporting a controller container's logs back to the Kurtosis initializer.

One common failure scenario we've seen with MacOS users is the `/var/folders` not being permitted in the Docker engine preferences. If you're a Mac user, double-check that your Docker engine's `Resources > File Sharing` section permits access to `/var/folders`.

If this still doesn't resolve the issue, you'll want to investigate the logs of your controller container, which you can do [using these instructions](https://docs.docker.com/config/containers/logging/); this should give you more information about why your container is failing.

Overlapping IP address ranges
-----------------------------
When Docker errors saying that subnet IP address ranges conflict, this usually means that Kurtosis is trying to create the per-test subnets but is colliding with an existing network that was left over from a previous invocation of the test suite. Kurtosis will clean up the Docker networks it creates under normal circumstances, but abnormal exits (e.g. SIGKILL) will leave the Docker networks hanging around. To fix this error, remove the offending networks like so:

```
docker network ls
docker network rm <ID of offending network>
```

Timeout while waiting for a service to start
--------------------------------------------
Before running a test against a network of services, Kurtosis performs availability checks on all the nodes in the network to ensure they're up. This is to avoid spurious test failures due to the network not being ready. If your test fails with an error like so:

```
Caused by: Hit timeout (1m30s) while waiting for service to start
 --- at /go/pkg/mod/github.com/kurtosis-tech/kurtosis@v0.0.0-20200722101726-28a51087d1db/commons/services/service_availability_checker.go:56 (ServiceAvailabilityChecker.WaitForStartup) ---
Caused by: context deadline exceeded
```

then the timeout Kurtosis is timing out while waiting for a node in a network to start. You should first examine why this might be the case to understand if there's a bug with your service or how you're checking for service availability, and if your service's slowness is expected then you can up the timeout that you define in your availability checker for your service.

Test execution timeout
----------------------
Each test has a timeout by which its execution must complete. A test's execution is specifically for the logic inside each test's `Run` method, and does NOT include the network setup required to even launch the test. A test's execution timeout being hit means that the test logic is not completing within the allowed deadline; the fix is to discover why the test is hitting the timeout (it may be a useful indicator of a problem) and, if no problems are found, increase the execution timeout in the test's `GetExecutionTimeout` method.

Hard test timeout
-----------------
In addition to the test execution timeout and in order to prevent any test from hanging forever, the entire test - including network setup, test execution, and network teardown - are subject to an additional "hard test timeout". This timeout is equal to the test execution timeout (configured in `GetExecutionTimeout`) plus the setup buffer (configured in `GetSetupBuffer`). If your test is hitting the hard test timeout but NOT the execution timeout, it likely means that some element of network setup is taking longer than expected. The initializers and availability checkers for your services should be examined for problems as a first step, and - if no issues are found - then the last fix should be increasing the setup buffer.

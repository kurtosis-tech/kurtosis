/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_container_mountpoints

/*
The dirpath on the TESTSUITE container where the execution volume will be mounted

This is hardcoded because the only way to get information back from teh test suite container is with a mounted file, but
	we won't know where to mount the file without getting information back from the test suite about where's an acceptable
	location to mount. This gives us a catch-22, which we solve by hardcoding the mount location and forcing the user
	to make it a mountable location in their image.
 */
const (
	TestsuiteContainerSuiteExVolMountpoint = "/suite-execution"
)

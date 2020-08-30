/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_mount_locations

/*
These filepaths are hardcoded locations where bind mounts and volumes will be mounted on the test suite image

These are hardcoded because without bind-mounting or making the test suite container run a server (which we definitely
	don't want to do), there's no way to get information back from the test suite image. This gives us a catch-22:
	we'd need to get information out of the test suite image to make the bind mount location configurable, but we need
	to bind mount in order to get information from the test suite image.

To avoid all that, we just hardcode the bind mount location and we can deal with it later if it becomes problematic.
 */
const (
	// Where bind
	BindMountsDirpath = "/bind-mounts"

	TestVolumeDirpath = "/test-volume"
)

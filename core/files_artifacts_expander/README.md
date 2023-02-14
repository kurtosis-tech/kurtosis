Files Artifacts Expander
========================
Files artifacts are packages of files stored on the API container's enclave data directory that the user can mount onto services they start at arbitrary locations.

The "mount files at an arbitrary location" is accomplished by volumes, so that if we have files in a volume then we can mount that volume at an arbitrary location on the user service.

The files artifacts are stored in `.tgz` form, meaning that we need to expand the archive into a volume before we mount them on the user service.

This is the function of the Docker image produced by this subproject: provide an image that the API container can run to decompress files artifacts into volumes before the user service starts.

Under the hood, this image works by a) downloading a files artifact from the API container using the API containe's `DownloadFilesArtifact` endpoint and 2) un-`tar`ing that file into a specified location which will just happen to be a volume usable by the user service.

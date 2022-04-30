package local_static_file_test

/*
const (
	testName = "upload-files"
	isPartitioningEnabled = false

	fileServerServiceImage                          = "flashspys/nginx-static"
	fileServerServiceId          services.ServiceID = "file-server"
	fileServerPortId = "http"
	fileServerPrivatePortNum = 80

	waitForStartupTimeBetweenPolls = 500
	waitForStartupMaxRetries       = 15
	waitInitialDelayMilliseconds   = 0

	testFileContents = "These are test file contents"

	userServiceMountPointForTestFilesArtifact = "/static"
)
var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
)

func TestStoreWebFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifactId, generatedFilename, err := storeGeneratedFile(t, enclaveCtx)
	require.NoError(t, err, "An error occurred storing the files artifact")

	filesArtifactMountpoints := map[services.FilesArtifactID]string{
		filesArtifactId: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier(filesArtifactMountpoints)

	serviceCtx, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")
	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	require.True(t, found, "Expected to find public port for ID '%v', but none was found", fileServerPortId)
	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()

	require.NoError(t,
		// TODO It's suuuuuuuuuuper confusing that we have to pass the private port in here!!!! We should just require the user
		//  to pass in the port ID and the API container will translate that to the private port automatically!!!
		enclaveCtx.WaitForHttpGetEndpointAvailability(fileServerServiceId, fileServerPrivatePortNum, generatedFilename, waitInitialDelayMilliseconds, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls, ""),
		"An error occurred waiting for the file server service to become available",
	)
	logrus.Infof("Added file server service with public IP '%v' and port '%v'", fileServerPublicIp, fileServerPublicPortNum)

	// ------------------------------------- TEST RUN ----------------------------------------------
	file1Contents, err := getFileContents(
		fileServerPublicIp,
		fileServerPublicPortNum,
		generatedFilename,
	)
	require.NoError(t, err, "An error occurred getting file 1's contents")
	require.Equal(
		t,
		testFileContents,
		file1Contents,
		"Actual file 1 contents '%v' != expected file 1 contents '%v'",
		file1Contents,
		testFileContents,
	)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getFileServerContainerConfigSupplier(filesArtifactMountpoints map[services.FilesArtifactID]string) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			fileServerServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			fileServerPortId: fileServerPortSpec,
		}).WithFiles(
			filesArtifactMountpoints,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getFileContents(ipAddress string, portNum uint16, filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, portNum, filename))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", filename)
	}
	body := resp.Body
	defer func() {
		if err := body.Close(); err != nil {
			logrus.Warnf("We tried to close the response body, but doing so threw an error:\n%v", err)
		}
	}()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}









const (
	testName = "files"
	isPartitioningEnabled = false

	dockerImage                    = "alpine:3.12.4"
	testService services.ServiceID = "test-service"

	testFileContents = "This is a test file"
	mountpointOnContainer = "/artifact-contents"

	generatedFilePermBits = 0644

	execCommandSuccessExitCode = int32(0)

)

func TestFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()


	// ------------------------------------- TEST SETUP ----------------------------------------------
	tempDirpath, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	filepath := path.Join(tempDirpath, testFilename)
	require.NoError(t, ioutil.WriteFile(filepath, []byte(testFileContents), generatedFilePermBits))

	// ------------------------------------- TEST RUN ----------------------------------------------
	filesArtifactId, generatedFilename, err := storeGeneratedFile(t, enclaveCtx)
	require.NoError(t, err)

	containerConfigSupplier := getContainerConfigSupplier(filesArtifactId)
	serviceCtx, err := enclaveCtx.AddService(testService, containerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")

	for relativeFilepath, expectedContents := range generatedFileRelPathsAndContents {
		sharedFilepath := serviceCtx.GetSharedDirectory().GetChildPath(relativeFilepath)

		catStaticFileCmd := []string{
			"cat",
			sharedFilepath.GetAbsPathOnServiceContainer(),
		}

		exitCode, logOutput, err := serviceCtx.ExecCommand(catStaticFileCmd)
		require.NoError(t, err, "An error occurred executing command '%+v' to cat the static file '%v' contents", catStaticFileCmd, relativeFilepath)
		require.Equal(
			t,
			execCommandSuccessExitCode,
			exitCode,
			"Command '%+v' to cat the static file '%v' exited with non-successful exit code '%v'",
			catStaticFileCmd,
			relativeFilepath,
			exitCode,
		)
		actualContents := logOutput
		require.Equal(
			t,
			expectedContents,
			actualContents,
			"Static file contents '%v' don't match expected static file '%v' contents '%v'",
			actualContents,
			relativeFilepath,
			expectedContents,
		)
		logrus.Infof("Static file '%v' contents were '%v' as expected", relativeFilepath, expectedContents)
	}
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getContainerConfigSupplier(filesArtifactId services.FilesArtifactID) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string) (*services.ContainerConfig, error) {
		// We sleep because the only function of this container is to test Docker executing a command while it's running
		// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
		// args), but this provides a nice little regression test of the ENTRYPOINT overriding
		entrypointArgs := []string{"sleep"}
		cmdArgs := []string{"30"}

		containerConfig := services.NewContainerConfigBuilder(
			dockerImage,
		).WithEntrypointOverride(
			entrypointArgs,
		).WithCmdOverride(
			cmdArgs,
		).WithFiles(map[services.FilesArtifactID]string{
			filesArtifactId: mountpointOnContainer,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}


func storeGeneratedFile(t *testing.T, enclaveCtx *enclaves.EnclaveContext) (
	resultArtifactId services.FilesArtifactID,
	storedFilename string,
	resultErr error,
) {
	tempFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer tempFile.Close()

	_, err = tempFile.WriteString(testFileContents)
	require.NoError(t, err)

	artifactId, err := enclaveCtx.UploadFiles(tempFile.Name())
	require.NoError(t, err)

	return artifactId, path.Base(tempFile.Name()), nil
}



 */
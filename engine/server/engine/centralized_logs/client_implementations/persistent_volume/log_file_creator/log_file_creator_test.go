package log_file_creator

const (
	testEnclaveUuid = "test-enclave-1"
)

//func TestLogFileCreatorCreatesNewFilePaths(t *testing.T) {
//	// mock the filesystem
//	mapFS := &fstest.MapFS{}
//	fs := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
//
//	// mock the time
//	time := logs_clock.NewMockLogsClock(2023, 10, 5)
//
//	// mock kurtosis backend GetUserServices call
//	kurtosisBackend := backend_interface.NewMockKurtosisBackend(t)
//	//kurtosisBackend.EXPECT().
//	//	GetUserService()
//
//	expectedFilePaths := []string{}
//
//	// execute create new file paths function
//	ctx := context.Background()
//	fileCreator := NewLogFileCreator(time, kurtosisBackend, fs)
//	err := fileCreator.CreateLogFiles(ctx)
//
//	require.NoError(t, err)
//	// check they exist in the mocked filesystem
//	for _, path := range expectedFilePaths {
//		file, err := fs.Stat(path)
//		require.NoError(t, err)
//		require.NotNil(t, file)
//	}
//}
//
//func TestLogFileCreatorCreatesFilePathsIdempotently(t *testing.T) {
//	// mock the filesystem
//	// add file paths to the file system
//
//	// mock the time
//
//	// mock kurtosis backend GetUserServices call
//
//	// execute create new file paths function
//
//	// check that no new files are added
//}
//
//func TestLogFileCreatorCreatesNewFilePathsForOnlyNonExistentFilePaths(t *testing.T) {
//
//}

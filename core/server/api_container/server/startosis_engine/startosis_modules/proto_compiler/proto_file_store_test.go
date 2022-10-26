package proto_compiler

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/mock_module_content_provider"
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

func TestLoadProtoFile_FullComputationValidProto(t *testing.T) {
	protoFileInModule := "github.com/kurtosis/module/types.proto"
	protoFileContent := `syntax = "proto3";

message InputArgs {
  string hello = 1; // world !
}
`

	mockModuleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer mockModuleContentProvider.RemoveAll()
	require.Nil(t, mockModuleContentProvider.AddFileContent(protoFileInModule, protoFileContent))
	protoFileAbsPath, err := mockModuleContentProvider.GetOnDiskAbsoluteFilePath(protoFileInModule)
	require.Nil(t, err)

	store := NewProtoFileStore(mockModuleContentProvider)
	protoRegistryFile, err := store.LoadProtoFile(protoFileInModule)
	require.Nil(t, err)

	// check that InputArgs file descriptor is loaded
	inputArgsDescriptor, err := protoRegistryFile.FindDescriptorByName("InputArgs")
	require.Nil(t, err)
	require.NotNil(t, inputArgsDescriptor)

	// check that the result has been stored
	fileUniqueIdentifier, err := getFileUniqueIdentifier(protoFileAbsPath)
	require.Nil(t, err)
	storedProtoRegistryFile, found := store.store[fileUniqueIdentifier]
	require.True(t, found)
	storedInputArgsDescriptor, err := storedProtoRegistryFile.FindDescriptorByName("InputArgs")
	require.Nil(t, err)
	require.NotNil(t, storedInputArgsDescriptor)
}

func TestLoadProtoFile_FullComputation_InvalidValidProtoWithClearCompilationError(t *testing.T) {
	protoFileInModule := "github.com/kurtosis/module/types.proto"
	protoFileContent := `syntax = "proto3";

message InputArgs {
  string hello = 1 // missing ';' here
}
`

	mockModuleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer mockModuleContentProvider.RemoveAll()
	require.Nil(t, mockModuleContentProvider.AddFileContent(protoFileInModule, protoFileContent))
	protoFileAbsPath, err := mockModuleContentProvider.GetOnDiskAbsoluteFilePath(protoFileInModule)
	require.Nil(t, err)

	store := NewProtoFileStore(mockModuleContentProvider)
	protoRegistryFile, err := store.LoadProtoFile(protoFileInModule)
	require.Nil(t, protoRegistryFile)
	require.NotNil(t, err)

	expectedErrorMessageTemplate := `
Caused by: Unable to compile .proto file '%s' (checked out at '%s'). Proto compiler output was: 
%s:5:1: Expected ";".
`
	partOfExpectedErrorMessage := fmt.Sprintf(expectedErrorMessageTemplate, protoFileInModule, protoFileAbsPath, path.Base(protoFileAbsPath))
	require.Contains(t, err.Error(), partOfExpectedErrorMessage)

	// check that nothing was stored
	fileUniqueIdentifier, err := getFileUniqueIdentifier(protoFileAbsPath)
	require.Nil(t, err)
	storedProtoRegistryFile, found := store.store[fileUniqueIdentifier]
	require.False(t, found)
	require.Nil(t, storedProtoRegistryFile)
}

func TestLoadProtoFile_FullComputation_FileContentUpdateIsReloaded(t *testing.T) {
	protoFileInModule := "github.com/kurtosis/module/types.proto"
	protoFileContent := `syntax = "proto3";

message InputAgrs { // With a Typo!
  string hello = 1;
}
`

	mockModuleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer mockModuleContentProvider.RemoveAll()
	require.Nil(t, mockModuleContentProvider.AddFileContent(protoFileInModule, protoFileContent))

	store := NewProtoFileStore(mockModuleContentProvider)
	protoRegistryFile, err := store.LoadProtoFile(protoFileInModule)
	require.Nil(t, err)

	// check that InputArgs file descriptor is loaded
	inputArgsDescriptor, err := protoRegistryFile.FindDescriptorByName("InputAgrs")
	require.Nil(t, err)
	require.NotNil(t, inputArgsDescriptor)

	// Update the content of the proto file and reload it
	correctedProtoFileContent := `syntax = "proto3";

message InputArgs {
  string hello = 1;
}
`
	require.Nil(t, mockModuleContentProvider.AddFileContent(protoFileInModule, correctedProtoFileContent))
	newProtoRegistryFile, err := store.LoadProtoFile(protoFileInModule)
	require.Nil(t, err)

	// check that InputAgrs has disappeared and been replaced with the correct InputArgs
	inputArgsDescriptorWithTypo, err := newProtoRegistryFile.FindDescriptorByName("InputAgrs")
	require.NotNil(t, err)
	require.Nil(t, inputArgsDescriptorWithTypo)
	correctedArgsDescriptor, err := newProtoRegistryFile.FindDescriptorByName("InputArgs")
	require.Nil(t, err)
	require.NotNil(t, correctedArgsDescriptor)
}

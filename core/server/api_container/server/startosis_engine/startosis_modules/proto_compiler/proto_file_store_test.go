package proto_compiler

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/mock_module_content_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path"
	"strings"
	"testing"
)

func TestLoadProtoFile_FullComputationValidProto(t *testing.T) {
	protoFileInModule := "github.com/kurtosis/module/types.proto"
	protoFileContent := `syntax = "proto3";

message InputArgs {
  string hello = 1; // world !
}
`

	mockModuleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(map[string]string{
		protoFileInModule: protoFileContent,
	})
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
	storeKey, storedValue, err := store.getStoredEntryOrNil(protoFileAbsPath, protoFileInModule)
	require.Nil(t, err)
	require.True(t, strings.HasPrefix(string(storeKey), fmt.Sprintf("%s___", protoFileAbsPath)))
	storedInputArgsDescriptor, err := storedValue.FindDescriptorByName("InputArgs")
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

	mockModuleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(map[string]string{
		protoFileInModule: protoFileContent,
	})
	protoFileAbsPath, err := mockModuleContentProvider.GetOnDiskAbsoluteFilePath(protoFileInModule)
	require.Nil(t, err)

	store := NewProtoFileStore(mockModuleContentProvider)
	protoRegistryFile, err := store.LoadProtoFile(protoFileInModule)
	require.Nil(t, protoRegistryFile)
	require.NotNil(t, err)
	expectedErrorMessageContains := fmt.Sprintf(`
Caused by: Unable to compile .proto file '%s' (checked out at '%s'). Proto compiler output was: 
%s:5:1: Expected ";".
`, protoFileInModule, protoFileAbsPath, path.Base(protoFileAbsPath))

	assert.Contains(t, err.Error(), expectedErrorMessageContains)

	// check that nothing was stored
	storeKey, storedValue, err := store.getStoredEntryOrNil(protoFileAbsPath, protoFileInModule)
	require.Nil(t, err)
	require.True(t, strings.HasPrefix(string(storeKey), fmt.Sprintf("%s___", protoFileAbsPath)))
	require.Nil(t, storedValue)
}

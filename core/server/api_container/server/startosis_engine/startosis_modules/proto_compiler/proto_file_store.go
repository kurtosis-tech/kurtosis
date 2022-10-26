package proto_compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"os"
	"os/exec"
	"path"
	"sync"
)

const (
	defaultTempDir             = ""
	protocTempOutputDirPattern = "protobuf-compiled-file-*.pb"

	storeKeyTemplate = "%s___%s"
)

var (
	protoUnmarshalerOptions = proto.UnmarshalOptions{Merge: true}
)

type StoreKey string

type ProtoFileStore struct {
	moduleProvider startosis_modules.ModuleContentProvider

	// Stores the compiled protoregistry.Files for each proto file. The key is a composite
	// <proto_file_path_on_disk, proto_file_hash> to guard against the proto file content changing when the module is
	// cloned again.
	store map[StoreKey]*protoregistry.Files

	// Use a mutex to avoid loading a file twice b/c the second load had happened before the first one finished.
	// For now, single common mutex for all files.
	// If it becomes the bottleneck, we can easily do one mutex per file in the store map
	mutex *sync.Mutex
}

func NewProtoFileStore(moduleProvider startosis_modules.ModuleContentProvider) *ProtoFileStore {
	return &ProtoFileStore{
		mutex:          &sync.Mutex{},
		moduleProvider: moduleProvider,
		store:          make(map[StoreKey]*protoregistry.Files),
	}
}

func (protoStore *ProtoFileStore) LoadProtoFile(protoModuleFile string) (*protoregistry.Files, error) {
	protoStore.mutex.Lock()
	defer protoStore.mutex.Unlock()

	// Get the path of the corresponding file on disk from the module provider
	absProtoFileOnDiskPath, err := protoStore.moduleProvider.GetOnDiskAbsoluteFilePath(protoModuleFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed loading the proto file '%v' from the module provider", protoModuleFile)
	}

	// Check in the store in case we already compiled it
	protoFileUniqueIdentifier, err := getFileUniqueIdentifier(absProtoFileOnDiskPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error computing unique file identifier for .proto file '%v' (checked out at '%v')", protoModuleFile, absProtoFileOnDiskPath)
	}
	if protoTypesRegistry, found := protoStore.store[protoFileUniqueIdentifier]; found {
		return protoTypesRegistry, nil
	}

	// Compile and load the protobuf types
	compiledProtoFileContent, err := compileProtoFile(absProtoFileOnDiskPath, protoModuleFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to compile the .proto file '%v' (checked out at '%v') using protobuf compiler", protoModuleFile, absProtoFileOnDiskPath)
	}

	protoTypesRegistry, err := loadTypesFromCompiledProtoIntoRegistry(compiledProtoFileContent, protoModuleFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to load types from the protobuf types registry generated from file '%v' (checked out at '%v')", protoModuleFile, absProtoFileOnDiskPath)
	}

	// Store for potential future calls and return
	protoStore.store[protoFileUniqueIdentifier] = protoTypesRegistry
	return protoTypesRegistry, nil
}

func compileProtoFile(absProtoFileOnDiskPath string, protoModuleFileForLogging string) ([]byte, error) {
	tmpCompiledProtobufFile, err := os.CreateTemp(defaultTempDir, protocTempOutputDirPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to create a temporary folder on disk to store the protoc output files")
	}
	defer os.RemoveAll(tmpCompiledProtobufFile.Name())
	absProtoFileDirPath := path.Dir(absProtoFileOnDiskPath)
	compileProtoCommand := exec.Command("protoc", "-I="+absProtoFileDirPath, "--descriptor_set_out="+tmpCompiledProtobufFile.Name(), absProtoFileOnDiskPath)

	if cmdOutput, err := compileProtoCommand.CombinedOutput(); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to compile .proto file '%s' (checked out at '%v'). Proto compiler output was: \n%v", protoModuleFileForLogging, absProtoFileOnDiskPath, string(cmdOutput))
	}

	compiledProtobufFileContent, err := os.ReadFile(tmpCompiledProtobufFile.Name())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read content of compiled .proto file '%v' (checked out at '%v' and compiled at '%v')", protoModuleFileForLogging, absProtoFileOnDiskPath, tmpCompiledProtobufFile.Name())
	}
	return compiledProtobufFileContent, nil
}

func loadTypesFromCompiledProtoIntoRegistry(compiledProtoFileContent []byte, protoModuleFileForLogging string) (*protoregistry.Files, error) {
	var protoFileDescriptorSet descriptorpb.FileDescriptorSet
	if err := protoUnmarshalerOptions.Unmarshal(compiledProtoFileContent, &protoFileDescriptorSet); err != nil {
		return nil, stacktrace.Propagate(err, "Unable read content of compiled .proto file '%v' and convert it to a file descriptor set", protoModuleFileForLogging)
	}

	protoRegistryFiles, err := protodesc.NewFiles(&protoFileDescriptorSet)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to convert proto file '%v' to a proto registry file set", protoModuleFileForLogging)
	}
	return protoRegistryFiles, nil
}

func getFileUniqueIdentifier(absProtoFileOnDiskPath string) (StoreKey, error) {
	fileContent, err := os.ReadFile(absProtoFileOnDiskPath)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to read file content '%v'", absProtoFileOnDiskPath)
	}

	fileHash := sha256.Sum256(fileContent)
	fileHashStr := hex.EncodeToString(fileHash[:])
	return StoreKey(fmt.Sprintf(storeKeyTemplate, absProtoFileOnDiskPath, fileHashStr)), nil
}

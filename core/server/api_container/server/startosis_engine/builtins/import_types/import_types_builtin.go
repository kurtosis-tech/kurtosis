package import_types

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/proto_compiler"
	starlarkproto "go.starlark.net/lib/proto"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	ImportTypesBuiltinName = "import_types"
)

func GenerateImportTypesBuiltin(protoFileStore *proto_compiler.ProtoFileStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var fileInModuleId string
		if err := starlark.UnpackArgs(ImportTypesBuiltinName, args, kwargs, "types_file", &fileInModuleId); err != nil {
			return nil, startosis_errors.NewInterpretationError("Unable to parse arguments of command " + ImportTypesBuiltinName + ". It should be a single string argument pointing to the types file to import")
		}

		protoRegistryFile, err := protoFileStore.LoadProtoFile(fileInModuleId)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Unable to load types file " + fileInModuleId + ". Error was: " + err.Error())
		}

		typesStringDict := loadTypesFromProtoRegistry(protoRegistryFile)
		return starlarkstruct.FromStringDict(starlarkstruct.Default, *typesStringDict), nil
	}
}

func loadTypesFromProtoRegistry(protoTypesFiles *protoregistry.Files) *starlark.StringDict {
	typesStringDict := starlark.StringDict{}
	protoTypesFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		starlarkFileDescriptor := &starlarkproto.FileDescriptor{
			Desc: fileDescriptor,
		}
		for _, typeName := range starlarkFileDescriptor.AttrNames() {
			typeValue, _ := starlarkFileDescriptor.Attr(typeName)
			typesStringDict[typeName] = typeValue
		}
		return true // true to continue iteration
	})
	return &typesStringDict
}

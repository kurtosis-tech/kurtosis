package import_types

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/proto_compiler"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	starlarkproto "go.starlark.net/lib/proto"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"strings"
)

const (
	ImportTypesBuiltinName = "import_types"

	typesFileArgName = "types_file"
)

func GenerateImportTypesBuiltin(protoFileStore *proto_compiler.ProtoFileStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var fileInModuleId string
		if err := starlark.UnpackArgs(ImportTypesBuiltinName, args, kwargs, typesFileArgName, &fileInModuleId); err != nil {
			return nil, startosis_errors.NewInterpretationError("Unable to parse arguments of command " + ImportTypesBuiltinName + ". It should be a single string argument pointing to the fully qualified .proto types file (i.e. \"github.com/kurtosis/module/types.proto\")")
		}

		protoRegistryFile, err := protoFileStore.LoadProtoFile(fileInModuleId)
		if err != nil {
			logrus.Errorf(stacktrace.Propagate(err, "Unable to load types file '%v'", fileInModuleId).Error())
			return nil, startosis_errors.NewInterpretationError("Unable to load types file " + fileInModuleId + ". Is the corresponding type file present in the module?")
		}

		typesStringDict, errorsMap := loadTypesFromProtoRegistry(protoRegistryFile)
		if len(errorsMap) != 0 {
			var failedTypes []string
			for failingTypeName := range errorsMap {
				failedTypes = append(failedTypes, failingTypeName)
			}
			logrus.Errorf("Error loading types for module '%s' from the proto registry. Errors were: '%v'", fileInModuleId, errorsMap)
			return nil, startosis_errors.NewInterpretationError("Unable to load types file " + fileInModuleId + ". The following types could not be loaded from the types registry: " + strings.Join(failedTypes, ", "))
		}
		return starlarkstruct.FromStringDict(starlarkstruct.Default, *typesStringDict), nil
	}
}

func loadTypesFromProtoRegistry(protoTypesFiles *protoregistry.Files) (*starlark.StringDict, map[string]error) {
	typesStringDict := starlark.StringDict{}
	errors := make(map[string]error)
	protoTypesFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		starlarkFileDescriptor := &starlarkproto.FileDescriptor{
			Desc: fileDescriptor,
		}
		for _, typeName := range starlarkFileDescriptor.AttrNames() {
			typeValue, err := starlarkFileDescriptor.Attr(typeName)
			if err != nil {
				errors[typeName] = stacktrace.Propagate(err, "Unable to load type '%s' from the proto registry", typeName)
			}
			if typeValue == nil {
				errors[typeName] = stacktrace.NewError("Unable to load type '%s' from the proto registry", typeName)
			}
			typesStringDict[typeName] = typeValue
		}
		return true // true to continue iteration
	})
	if len(errors) != 0 {
		return nil, errors
	}
	return &typesStringDict, nil
}

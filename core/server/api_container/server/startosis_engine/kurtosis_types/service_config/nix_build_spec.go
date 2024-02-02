package service_config

import (
	"path"
	"path/filepath"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

const (
	NixBuildSpecTypeName = "NixBuildSpec"

	FlakeLocationDir = "flake_location_dir"
	FlakeOutputAttr  = "flake_output"
	NixContextAttr   = "build_context_dir"

	// Currently only supports container nixs named Dockerfile
	defaultNixFlakeFilePath = "flake.nix"
)

func NewNixBuildSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: NixBuildSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              FlakeLocationDir,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, FlakeLocationDir)
					},
				},
				{
					Name:              NixContextAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, NixContextAttr)
					},
				},
				{
					Name:              FlakeOutputAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, FlakeOutputAttr)
					},
				},
			},
		},
		Instantiate: instantiateNixBuildSpec,
	}
}

func instantiateNixBuildSpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(NixBuildSpecTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &NixBuildSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// NixBuildSpec is a starlark.Value that holds all the information for the startosis_engine to initiate an nix build
type NixBuildSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (nixBuildSpec *NixBuildSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := nixBuildSpec.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &NixBuildSpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// Relative locator of build context directory
func (nixBuildSpec *NixBuildSpec) GetBuildContextLocator() (string, *startosis_errors.InterpretationError) {
	buildContextLocator, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](nixBuildSpec.KurtosisValueTypeDefault, NixContextAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			NixContextAttr, NixBuildSpecTypeName)
	}
	buildContextLocatorStr := buildContextLocator.GoString()
	return buildContextLocatorStr, nil
}

func (nixBuildSpec *NixBuildSpec) GetFlakeOutput() (string, *startosis_errors.InterpretationError) {
	flakeOutput, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](nixBuildSpec.KurtosisValueTypeDefault, FlakeOutputAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", nil
	}
	return flakeOutput.GoString(), nil
}

func (nixBuildSpec *NixBuildSpec) GetFlakeLocationDir() (string, *startosis_errors.InterpretationError) {
	flakeLocator, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](nixBuildSpec.KurtosisValueTypeDefault, FlakeLocationDir)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", nil
	}
	return flakeLocator.GoString(), nil
}

func (nixBuildSpec *NixBuildSpec) ToKurtosisType(
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (*nix_build_spec.NixBuildSpec, *startosis_errors.InterpretationError) {
	// get locator of context directory (relative or absolute)
	buildContextLocator, interpretationErr := nixBuildSpec.GetBuildContextLocator()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	flakeLocationDir, interpretationErr := nixBuildSpec.GetFlakeLocationDir()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	buildContextDirPathOnDisk, flakeNixFilePathOnDisk, interpretationErr := getOnDiskNixBuildSpecPaths(
		buildContextLocator,
		flakeLocationDir,
		packageId,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageContentProvider,
		packageReplaceOptions)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	flakeOutputStr, interpretationErr := nixBuildSpec.GetFlakeOutput()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return nix_build_spec.NewNixBuildSpec(buildContextDirPathOnDisk, flakeNixFilePathOnDisk, flakeOutputStr), nil
}

// Returns the filepath of the build context directory and flake nix on APIC based on package info
func getOnDiskNixBuildSpecPaths(
	buildContextLocator string,
	flakeLocationDir string,
	packageId string,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (string, string, *startosis_errors.InterpretationError) {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript {
		return "", "", startosis_errors.NewInterpretationError("Cannot use NixBuildSpec in a standalone script; create a package and rerun to use NixBuildSpec.")
	}

	// get absolute locator of context directory
	contextDirAbsoluteLocator, interpretationErr := packageContentProvider.GetAbsoluteLocator(packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, buildContextLocator, packageReplaceOptions)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	// get on disk directory path of Dockerfile
	flakeNixAbsoluteLocator := path.Join(contextDirAbsoluteLocator, flakeLocationDir, defaultNixFlakeFilePath)

	flakeNixPathOnDisk, interpretationErr := packageContentProvider.GetOnDiskAbsolutePackageFilePath(flakeNixAbsoluteLocator)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	// Assume, that flake nix sits at the same level as context directory to get context dir path on disk
	contextDirPathOnDisk := filepath.Dir(flakeNixPathOnDisk)

	return contextDirPathOnDisk, flakeNixPathOnDisk, nil
}

package nix_build_spec

import (
	"encoding/json"

	"github.com/kurtosis-tech/stacktrace"
)

type NixBuildSpec struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateNixBuildSpec *privateNixBuildSpec
}

// NixBuildSpec contains the information need for building a container from nix.
type privateNixBuildSpec struct {
	NixFlakeDir    string
	ContextDirPath string
	FlakeOutput    string
}

func NewNixBuildSpec(contextDirPath string, nixFlakeDir string, flakeOutput string) *NixBuildSpec {
	internalNixBuildSpec := &privateNixBuildSpec{
		NixFlakeDir:    nixFlakeDir,
		ContextDirPath: contextDirPath,
		FlakeOutput:    flakeOutput,
	}
	return &NixBuildSpec{internalNixBuildSpec}
}

func (nixBuildSpec *NixBuildSpec) GetNixFlakeDir() string {
	return nixBuildSpec.privateNixBuildSpec.NixFlakeDir
}

func (nixBuildSpec *NixBuildSpec) GetBuildContextDir() string {
	return nixBuildSpec.privateNixBuildSpec.ContextDirPath
}

func (nixBuildSpec *NixBuildSpec) GetFlakeOutput() string {
	return nixBuildSpec.privateNixBuildSpec.FlakeOutput
}

func (nixBuildSpec *NixBuildSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(nixBuildSpec.privateNixBuildSpec)
}

func (nixBuildSpec *NixBuildSpec) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateNixBuildSpec{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	nixBuildSpec.privateNixBuildSpec = unmarshalledPrivateStructPtr
	return nil
}

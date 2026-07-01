package image_org

import "os"

// EnvVarName overrides the org/registry prefix of Kurtosis' own images (engine,
// core, files-artifacts-expander), e.g. "ghcr.io/my-org"; the engine propagates it to APICs.
const EnvVarName = "KURTOSIS_IMAGE_ORG"

const defaultOrg = "kurtosistech"

func Get() string {
	if org := os.Getenv(EnvVarName); org != "" {
		return org
	}
	return defaultOrg
}

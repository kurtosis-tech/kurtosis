package annotations

const (
	SkipKurtosisConfigInitializationOnGlobalSetupKey   = "skip-kurtosis-config-initialization-on-global-setup"
	SkipKurtosisConfigInitializationOnGlobalSetupValue = ""
)

func ShouldSkipKurtosisConfigInitializationOnGlobalSetup(annotations map[string]string) bool {
	if _, found := annotations[SkipKurtosisConfigInitializationOnGlobalSetupKey]; found {
		return true
	}
	return false
}

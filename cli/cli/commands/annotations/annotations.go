package annotations

const (
	SkipConfigInitializationOnGlobalSetupKey = "skip-config-on-global"
	SkipConfigInitializationOnGlobalSetupValue = ""
)

func ShouldSkipConfigInitializationOnGlobalSetup(annotations map[string]string) bool {
	if _, found := annotations[SkipConfigInitializationOnGlobalSetupKey]; found {
		return true
	}
	return false
}

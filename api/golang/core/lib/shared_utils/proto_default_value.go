package shared_utils

// GetOrDefaultBool extract the value of a protobuf `optional bool` field with a fallback value if absent
func GetOrDefaultBool(optionalValue *bool, defaultValue bool) bool {
	if optionalValue == nil {
		return defaultValue
	}
	return *optionalValue
}

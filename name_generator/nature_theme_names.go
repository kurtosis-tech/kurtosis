package name_generator

// GenerateNatureThemeNameForFileArtifacts - generates nature theme name for file artifacts
func GenerateNatureThemeNameForFileArtifacts() string {
	nouns := getNouns()
	adjectives := getAdjectives()

	args := generatorArgs{
		adjectives: adjectives,
		nouns:      nouns,
	}
	return generateNatureThemeNameForArtifactsInternal(args)
}

// this method creates or re-uses name generator struct to generate appropriate names
func generateNatureThemeNameForArtifactsInternal(args generatorArgs) string {
	nameGenerator := getNameGenerator()
	natureThemeName := nameGenerator.generateName(args)
	return natureThemeName
}

func convertMapSetToStringArray(data map[string]bool) []string {
	var arr []string
	for k, _ := range data {
		arr = append(arr, k)
	}
	return arr
}

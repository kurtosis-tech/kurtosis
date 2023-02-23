package name_generator

// GenerateNatureThemeNameForFileArtifacts - generates nature theme name for file artifacts
func GenerateNatureThemeNameForFileArtifacts() string {
	args := generatorArgs{
		adjectives: ADJECTIVES,
		nouns:      NOUNS,
	}

	nameGenerator := getNameGenerator()
	natureThemeName := nameGenerator.generateName(args)
	return natureThemeName
}

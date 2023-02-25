package name_generator

// GenerateNatureThemeNameForFileArtifacts - generates nature theme name for file artifacts
// This method hides the complexities of dealing with singleton pattern; the consumers do not
// need to worry about creating a struct etc. They can just expect to get a pseudo-random nature theme
func GenerateNatureThemeNameForFileArtifacts() string {
	args := generatorArgs{
		adjectives: ADJECTIVES,
		nouns:      FILE_ARTIFACT_NOUNS,
	}

	nameGenerator := getNameGenerator()
	natureThemeName := nameGenerator.generateName(args)
	return natureThemeName
}

// GenerateNatureThemeNameForEngine - generates nature theme name for file artifacts
// This method hides the complexities of dealing with singleton pattern; the consumers do not
// need to worry about creating a struct etc. They can just expect to get a pseudo-random nature theme
func GenerateNatureThemeNameForEngine() string {
	args := generatorArgs{
		adjectives: ADJECTIVES,
		nouns:      ENGINE_NOUNS,
	}

	nameGenerator := getNameGenerator()
	natureThemeName := nameGenerator.generateName(args)
	return natureThemeName
}

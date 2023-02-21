package name_generator

import "github.com/kurtosis-tech/kurtosis/common/name_generator/data_provider"

// GenerateNatureThemeNameForArtifacts - generate a nature theme name
// we read the data everytime we call this method without consistent ordering
// TODO: use singleton pattern to access data from data provider once
// TODO: add nature theme names for enclaves etc.
func GenerateNatureThemeNameForArtifacts() string {
	nouns := data_provider.GetNatureThemeNounsData()
	adjectives := data_provider.GetNatureThemeAdjectivesData()

	args := GeneratorArgs{
		adjectives: adjectives,
		nouns:      nouns,
	}
	return generateNatureThemeNameForArtifactsInternal(args)
}

func generateNatureThemeNameForArtifactsInternal(args GeneratorArgs) string {
	nameGenerator := GetNameGenerator()
	natureThemeName := nameGenerator.GenerateName(args)
	return natureThemeName
}

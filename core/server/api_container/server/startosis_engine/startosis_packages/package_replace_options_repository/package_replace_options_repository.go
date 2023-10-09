package package_replace_options_repository

type PackageReplaceOptionsRepository struct {
}

func NewPackageReplaceOptionsRepository() *PackageReplaceOptionsRepository {
	return &PackageReplaceOptionsRepository{}
}

func (repository *PackageReplaceOptionsRepository) Get(key string) map[string]string {
	panic("Implement me")
}

package v7

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

type KurtosisCloudConfigV7 struct {
	ApiUrl           *string `yaml:"api-url,omitempty"`
	Port             *uint   `yaml:"port,omitempty"`
	CertificateChain *string `yaml:"certificate-chain,omitempty"`
}

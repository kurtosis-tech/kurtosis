module github.com/kurtosis-tech/kurtosis/enclave-manager/local

go 1.20
replace (
	github.com/kurtosis-tech/kurtosis/enclave-manager => ../server
	github.com/kurtosis-tech/kurtosis/contexts-config-store => ../../contexts-config-store
	github.com/kurtosis-tech/kurtosis/kurtosis_version => ../../kurtosis_version
)

require (
	github.com/kurtosis-tech/kurtosis/enclave-manager v0.0.0-20230828153722-32770ca96513
  github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

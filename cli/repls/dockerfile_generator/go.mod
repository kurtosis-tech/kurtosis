module github.com/kurtosis-tech/kurtosis-cli/repl_dockerfile_generator

go 1.15

replace github.com/kurtosis-tech/kurtosis-cli/commons => ../../commons

require (
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220303202459-439ee9b1fc5a // indirect
	github.com/kurtosis-tech/kurtosis-cli/commons v0.0.0
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20220303202823-31e2fb3327ba
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
)

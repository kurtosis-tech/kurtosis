module github.com/kurtosis-tech/kurtosis-cli/repl_dockerfile_generator

go 1.15

replace github.com/kurtosis-tech/kurtosis-cli/commons => ../../commons

require (
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220308120154-5a34daa0625a // indirect
	github.com/kurtosis-tech/kurtosis-cli/commons v0.0.0
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20220407202732-4b688f8ec2ef
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
)

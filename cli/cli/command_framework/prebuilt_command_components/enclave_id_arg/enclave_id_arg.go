package enclave_id_arg

import "github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/args"

func NewEnclaveIDArg(
	// The arg key where this enclave ID argument will be stored
	argKey string,
	// We expect that the engine to be set up via the command's PreValidationAndRunFunc; this is the key where it's stored
	engineClientCtxKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {



	return &args.ArgConfig{
		Key:             "",
		IsOptional:      isOptional,
		DefaultValue:    nil,
		IsGreedy:        isGreedy,
		CompletionsFunc: nil,
		ValidationFunc:  nil,
	}
}

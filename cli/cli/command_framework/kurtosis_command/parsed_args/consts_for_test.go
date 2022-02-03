package parsed_args

import "github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command"

const (
	arg1Key = "arg1"
	arg2Key = "arg2"
	arg3Key = "arg3"

	arg1Value = "arg1Value"
	arg2Value = "arg2Value"
	arg3Value1 = "arg3Value1"
	arg3Value2 = "arg3Value2"
	arg3Value3 = "arg3Value3"

	flag1Key = "flag1"
	flag2Key = "flag2"
)

var validArgsConfig = []*kurtosis_command.ArgConfig{
	{
		Key: arg1Key,
	},
	{
		Key: arg2Key,
	},
	{
		Key:      arg3Key,
		IsGreedy: true,
	},
}
var validTokens = []string{
	arg1Value,
	arg2Value,
	arg3Value1,
	arg3Value2,
	arg3Value3,
}

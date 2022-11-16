/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

const (
	// WARNING: If you add new URLs here, make sure to add them to the urlsToValidateInTest below!!!
	Domain           = "kurtosistech.com"
	InfoEmail        = "inquiries@" + Domain
	DocumentationUrl = "https://docs." + Domain
	SupportEmail     = "support@" + Domain
	DiscordUrl       = "https://discord.gg/6Jjp9c89z9"
	GithubIssuesUrl  = "https://github.com/kurtosis-tech/kurtosis-testsuite-api-lib/issues"
	CLISetupDocsUrl  = DocumentationUrl + "/running-in-ci.html"

	//WARNING!! we shouldn't modify this constant 'allowedEnclaveIdCharsRegexStr' at least we already
	//upgrade the same one in the engine's server which is in this file 'engine/server/engine/enclave_manager/enclave_id/enclave_id_validator.go'
	AllowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`
)

// List of URLs whose validity will be verified in a test
var urlsToValidateInTest = []string{
	DocumentationUrl,
	DiscordUrl,
	GithubIssuesUrl,
	CLISetupDocsUrl,
}

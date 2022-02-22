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
	NewUserSignupUrl = "https://" + Domain + "/sign-up"
	CLISetupDocsUrl = DocumentationUrl + "/running-in-ci.html"
)

// List of URLs whose validity will be verified in a test
var urlsToValidateInTest = []string{
	DocumentationUrl,
	DiscordUrl,
	GithubIssuesUrl,
	NewUserSignupUrl,
	CLISetupDocsUrl,
}

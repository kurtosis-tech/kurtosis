/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

const (
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	//    If you add new URLs below, make sure to add them to the urlsToValidateInTest below!!!
	Domain                       = "kurtosis.com"
	InfoEmail                    = "inquiries@" + Domain
	DocumentationUrl             = "https://docs." + Domain
	SupportEmail                 = "support@" + Domain
	DiscordUrl                   = "https://discord.gg/6Jjp9c89z9"
	GithubIssuesUrl              = "https://github.com/kurtosis-tech/kurtosis-sdk"
	CLICommandsReferenceURL      = DocumentationUrl + "/cli"
	StarlarkPackagesReferenceURL = DocumentationUrl + "/reference/packages"
	StarlarkLocatorsReferenceURL = DocumentationUrl + "/reference/locators"
	//    If you add new URLs above, make sure to add them to the urlsToValidateInTest below!!!
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
)

// List of URLs whose validity will be verified in a test
var urlsToValidateInTest = []string{
	DocumentationUrl,
	DiscordUrl,
	GithubIssuesUrl,
	CLICommandsReferenceURL,
	StarlarkPackagesReferenceURL,
	StarlarkLocatorsReferenceURL,
}

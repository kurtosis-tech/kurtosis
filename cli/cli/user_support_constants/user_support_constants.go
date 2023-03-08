/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

const (
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	//    If you add new URLs below, make sure to add them to the urlsToValidateInTest below!!!
	Domain                       = "kurtosis.com"
	OldDomain                    = "kurtosistech.com" //This domain is still used for email accounts
	DocumentationUrl             = "https://docs." + Domain
	DiscordUrl                   = "https://discord.gg/6Jjp9c89z9"
	GithubIssuesUrl              = "https://github.com/kurtosis-tech/kurtosis/issues/new/choose"
	CLICommandsReferenceURL      = DocumentationUrl + "/cli"
	StarlarkPackagesReferenceURL = DocumentationUrl + "/reference/packages"
	StarlarkLocatorsReferenceURL = DocumentationUrl + "/reference/locators"
	UpgradeCLIInstructionsPage   = DocumentationUrl + "/install#upgrading"
	MetricsPhilosophyDocs        = DocumentationUrl + "/explanations/metrics-philosophy"
	HowImportWorksLink           = DocumentationUrl + "/explanations/how-do-kurtosis-imports-work"
	PackageDocLink               = DocumentationUrl + "/reference/packages"
	GoogleRe2SyntaxDocumentation = "https://github.com/google/re2/wiki/Syntax"
	KurtosisDiscordUrl           = "https://discord.com/channels/783719264308953108/783719264308953111"
	KurtosisOnBoardCalendlyUrl   = "https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding"
	FeedbackEmail                = "feedback@" + OldDomain
	FeedbackEmailLink            = "mailto:" + FeedbackEmail

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
	UpgradeCLIInstructionsPage,
	GoogleRe2SyntaxDocumentation,
	MetricsPhilosophyDocs,
	KurtosisDiscordUrl,
	KurtosisOnBoardCalendlyUrl,
	HowImportWorksLink,
	PackageDocLink,
}

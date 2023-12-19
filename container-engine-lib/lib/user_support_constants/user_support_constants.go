/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

const (
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	//    If you add new URLs below, make sure to add them to the urlsToValidateInTest below!!!
	Domain                            = "kurtosis.com"
	OldDomain                         = "kurtosistech.com" //This domain is still used for email accounts
	DocumentationUrl                  = "https://docs." + Domain
	GithubRepoUrl                     = "https://github.com/kurtosis-tech/kurtosis"
	GithubNewIssuesUrl                = GithubRepoUrl + "/issues/new"
	GitHubChooseNewIssuesUrl          = GithubNewIssuesUrl + "/choose"
	GitHubChooseNewIssuesUrlWitLabels = GitHubChooseNewIssuesUrl + "?labels="
	GitHubBugIssueUrl                 = GithubNewIssuesUrl + "?labels=bug&template=bug-report.yml"
	GitHubFeatureRequestIssueUrl      = GithubNewIssuesUrl + "?labels=feature+request&template=feature-request.yml"
	GitHubDocsIssueUrl                = GithubNewIssuesUrl + "?labels=docs&template=docs-issue.yml"
	CLICommandsReferenceURL           = DocumentationUrl + "/cli"
	StarlarkPackagesReferenceURL      = DocumentationUrl + "/advanced-concepts/packages"
	StarlarkLocatorsReferenceURL      = DocumentationUrl + "/advanced-concepts/locators"
	UpgradeCLIInstructionsPage        = DocumentationUrl + "/upgrade"
	MetricsPhilosophyDocs             = DocumentationUrl + "/advanced-concepts/metrics-philosophy"
	HowImportWorksLink                = DocumentationUrl + "/advanced-concepts/how-do-kurtosis-imports-work"
	PackageDocLink                    = DocumentationUrl + "/advanced-concepts/packages"
	GoogleRe2SyntaxDocumentation      = "https://github.com/google/re2/wiki/Syntax"
	KurtosisDiscordUrl                = "https://discord.gg/6Jjp9c89z9"
	KurtosisOnBoardCalendlyUrl        = "https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding"
	FeedbackEmail                     = "feedback@" + OldDomain
	FeedbackEmailLink                 = "mailto:" + FeedbackEmail
	KurtosisTechTwitterProfileLink    = "https://twitter.com/KurtosisTech"
	KurtosisCloudLink                 = "https://cloud." + Domain

	//    If you add new URLs above, make sure to add them to the urlsToValidateInTest below!!!
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
)

// List of URLs whose validity will be verified in a test
var urlsToValidateInTest = []string{
	DocumentationUrl,
	GithubRepoUrl,
	GithubNewIssuesUrl,
	GitHubChooseNewIssuesUrl,
	GitHubChooseNewIssuesUrlWitLabels,
	GitHubBugIssueUrl,
	GitHubFeatureRequestIssueUrl,
	CLICommandsReferenceURL,
	GitHubDocsIssueUrl,
	StarlarkPackagesReferenceURL,
	StarlarkLocatorsReferenceURL,
	UpgradeCLIInstructionsPage,
	GoogleRe2SyntaxDocumentation,
	MetricsPhilosophyDocs,
	KurtosisDiscordUrl,
	KurtosisOnBoardCalendlyUrl,
	HowImportWorksLink,
	KurtosisCloudLink,
}

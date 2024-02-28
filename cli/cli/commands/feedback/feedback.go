package feedback

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
)

var commandDescription = "Give feedback, file a bug report or feature request, or get help from the Kurtosis team. " +
	"See below for the many ways you can get in touch with us.\n\n" +
	"TIP: You can quickly type and send us feedback directly from the CLI. For example, " +
	"`kurtosis feedback \"I enjoy the enclave naming theme\"` will open the Kurtosis GitHub " +
	"\"Create New Issue\" page with the description pre-filled with \"I enjoy the enclave naming theme\".\n\n" +
	"This command can be used to deliver feedback to the Kurtosis team!\n\n" +
	"* Pass in your feedback as an argument using double quotations (e.g. kurtosis feedback \"my feedback\"), and press enter; this will open our GitHub Issues templates, pre-filled with your feedback/arg.\n" +
	"* Pass in the --email flag to open a draft email, pre-filled with your feedback/arg, to send to feedback@kurtosistech.com.\n" +
	"* Pass in the --calendly flag to open our Calendly link to schedule a 1:1 session with us for feedback and questions you may have!\n\n" +
	"See below for the some direct links as well.\n\n" +
	"* For bugs/issues, let us know in our GitHub " + user_support_constants.GitHubChooseNewIssuesUrl + "?version=" + kurtosis_version.GetVersion() + ".\n" +
	"* For general feedback, click here to email us " + user_support_constants.FeedbackEmailLink + ".\n" +
	"* If you need help getting started, schedule an on-boarding session with us on " + user_support_constants.KurtosisOnBoardCalendlyUrl + "."

const (
	commandShortDescription            = "Give feedback"
	emailFlagKey                       = "email"
	calendlyFlagKey                    = "calendly"
	bugFeedbackFlagKey                 = "bug"
	bugFeedbackShortFlagKey            = "b"
	featureRequestFeedbackFlagKey      = "feature"
	featureRequestFeedbackShortFlagKey = "f"
	docsFeedbackFlagKey                = "docs"
	docsFeedbackShortFlagKey           = "d"

	emailFlagUsageDescription     = "Opens your mail client to send us feedback via email."
	calendlyFlagDescription       = "When set, opens the link to our Calendly page to schedule a 1:1 session with a Kurtosis expert."
	bugFlagDescription            = "To specify that this is a bug feedback type"
	featureRequestFlagDescription = "To specify that this is a feature request feedback type"
	docsFlagDescription           = "To specify that this is a docs feedback type"

	bugEmailSubjectPrefix       = "[BUG]"
	featureRequestSubjectPrefix = "[FEATURE_REQUEST]"
	docsSubjectPrefix           = "[DOCS]"

	defaultOpenEmailLink    = "false"
	defaultOpenCalendlyLink = "false"
	defaultBug              = "false"
	defaultFeatureRequest   = "false"
	defaultDocs             = "false"

	userMsgArgKey          = "message"
	userMsgArgDefaultValue = "default-message-value"
	userMsgArgIsOptional   = true
	userMsgArgIsNotGreedy  = false

	flagKeysStrSeparator = ", "

	notDestinationTypeSelected = ""
	notFeedbackTypeSelected    = ""

	defaultDestinationType = "gitHub"
)

var FeedbackCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.FeedbackCmdStr,
	ShortDescription: commandShortDescription,
	LongDescription:  commandDescription,
	Flags: []*flags.FlagConfig{
		{
			Key:       emailFlagKey,
			Usage:     emailFlagUsageDescription,
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultOpenEmailLink,
		},
		{
			Key:       calendlyFlagKey,
			Usage:     calendlyFlagDescription,
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultOpenCalendlyLink,
		},
		{
			Key:       bugFeedbackFlagKey,
			Usage:     bugFlagDescription,
			Shorthand: bugFeedbackShortFlagKey,
			Type:      flags.FlagType_Bool,
			Default:   defaultBug,
		},
		{
			Key:       featureRequestFeedbackFlagKey,
			Usage:     featureRequestFlagDescription,
			Shorthand: featureRequestFeedbackShortFlagKey,
			Type:      flags.FlagType_Bool,
			Default:   defaultFeatureRequest,
		},
		{
			Key:       docsFeedbackFlagKey,
			Usage:     docsFlagDescription,
			Shorthand: docsFeedbackShortFlagKey,
			Type:      flags.FlagType_Bool,
			Default:   defaultDocs,
		},
	},
	Args: []*args.ArgConfig{
		{
			Key:            userMsgArgKey,
			DefaultValue:   userMsgArgDefaultValue,
			IsOptional:     userMsgArgIsOptional,
			IsGreedy:       userMsgArgIsNotGreedy,
			ValidationFunc: validateUserMsgArg,
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	// Args parsing and validation
	userMsg, err := getUserMsgFromArgs(args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while extracting the user message from args '%+v'", args)
	}

	selectedDestinationType, err := getSelectedDestinationType(flags)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting the selected destination type from flags '%+v'",
			flags,
		)
	}

	mutuallyExclusiveFeedbackTypeFlagKeys := []string{bugFeedbackFlagKey, featureRequestFeedbackFlagKey, docsFeedbackFlagKey}

	selectedFeedbackType, err := validateMutuallyExclusiveBooleanFlagsAndGetSelectedKey(flags, mutuallyExclusiveFeedbackTypeFlagKeys)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating mutually exclusive flags '%+v'",
			mutuallyExclusiveFeedbackTypeFlagKeys,
		)
	}

	gitHubIssueURL, err := getGitHubIssueURL(selectedFeedbackType, userMsg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting GitHub Issue URL")
	}

	if selectedDestinationType == defaultDestinationType {
		if err := multi_os_command_executor.OpenFile(gitHubIssueURL); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis GitHub issue page")
		}
		return nil
	}

	if selectedDestinationType == emailFlagKey {
		emailLinkWithSubjectAndBody, err := getEmailLinkWithSubjectAndBody(userMsg, selectedFeedbackType)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting email link with subject and body")
		}

		if err := multi_os_command_executor.OpenFile(emailLinkWithSubjectAndBody); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the feedback email link")
		}
		return nil
	}

	if selectedDestinationType == calendlyFlagKey {
		if err := multi_os_command_executor.OpenFile(user_support_constants.KurtosisOnBoardCalendlyUrl); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Calendly email link")
		}
	}

	return nil
}

func getUserMsgFromArgs(args *args.ParsedArgs) (string, error) {
	userMsg, err := args.GetNonGreedyArg(userMsgArgKey)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting the user message argument using flag key '%v'",
			userMsgArgKey,
		)
	}
	if userMsg == userMsgArgDefaultValue {
		userMsg = ""
	}
	userEncodedMsg := &url.URL{ //nolint:exhaustruct
		Path:        userMsg,
		Scheme:      "",
		Opaque:      "",
		User:        nil,
		Host:        "",
		RawPath:     "",
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	return userEncodedMsg.String(), nil
}

func getEmailLinkWithSubjectAndBody(prefilledBody string, selectedFeedbackType string) (string, error) {
	var prefilledSubject string
	switch selectedFeedbackType {
	case bugFeedbackFlagKey:
		prefilledSubject = bugEmailSubjectPrefix
	case featureRequestFeedbackFlagKey:
		prefilledSubject = featureRequestSubjectPrefix
	case docsFeedbackFlagKey:
		prefilledSubject = docsSubjectPrefix
	default:
		prefilledSubject = ""
	}

	emailLink := fmt.Sprintf(
		"%s?subject=%s&body=%s",
		user_support_constants.FeedbackEmailLink,
		prefilledSubject,
		prefilledBody,
	)

	return emailLink, nil
}

func validateUserMsgArg(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	userMsgArg, err := args.GetNonGreedyArg(userMsgArgKey)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting the user message arguments using flag key '%v'",
			userMsgArgKey,
		)
	}

	if userMsgArg == "" {
		return stacktrace.NewError("Error validating the user message argument, it can not be an empty string")
	}
	return nil
}

func getGitHubIssueURL(
	selectedFeedbackType string,
	userEncodedMsgStr string,
) (string, error) {

	var gitHubIssueBaseUrl string
	switch selectedFeedbackType {
	case bugFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GitHubBugIssueUrl
	case featureRequestFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GitHubFeatureRequestIssueUrl
	case docsFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GitHubDocsIssueUrl
	case notFeedbackTypeSelected:
		gitHubIssueBaseUrl = user_support_constants.GitHubChooseNewIssuesUrlWitLabels
	default:
		return "", stacktrace.NewError(
			"An error occurred while setting the GitHub URL, expected "+
				"to receive a know feedback type but '%v' was received instead; this is a bug in Kurtosis",
			selectedFeedbackType,
		)
	}

	gitHubIssueURL := fmt.Sprintf(
		"%s&version=%v&description=%s&background-and-motivation=%s",
		gitHubIssueBaseUrl,
		kurtosis_version.GetVersion(),
		userEncodedMsgStr,
		userEncodedMsgStr,
	)

	return gitHubIssueURL, nil
}

// TODO we could add this to the command framework
func validateMutuallyExclusiveBooleanFlagsAndGetSelectedKey(flags *flags.ParsedFlags, flagKeys []string) (string, error) {
	anyPreviousFlagSet := false
	selectedKey := ""
	for _, flagKey := range flagKeys {
		flagValue, err := flags.GetBool(flagKey)
		if err != nil {
			return "", stacktrace.Propagate(
				err,
				"Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!",
				flagValue,
			)
		}
		if flagValue && anyPreviousFlagSet {
			flagKeysStr := strings.Join(flagKeys, flagKeysStrSeparator)
			return "", stacktrace.NewError("Flags '%s' are mutually exclusive, you have to pass only one of them", flagKeysStr)
		}
		if flagValue {
			selectedKey = flagKey
			anyPreviousFlagSet = true
		}
	}
	return selectedKey, nil
}

func getSelectedDestinationType(
	flags *flags.ParsedFlags,
) (string, error) {

	mutuallyExclusiveFeedbackDestinationTypeFlagKeys := []string{
		emailFlagKey, calendlyFlagKey}

	selectedDestinationType, err := validateMutuallyExclusiveBooleanFlagsAndGetSelectedKey(
		flags,
		mutuallyExclusiveFeedbackDestinationTypeFlagKeys,
	)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred validating mutually exclusive flags '%+v'",
			mutuallyExclusiveFeedbackDestinationTypeFlagKeys,
		)
	}

	if selectedDestinationType == notDestinationTypeSelected {
		selectedDestinationType = defaultDestinationType
	}

	return selectedDestinationType, nil
}

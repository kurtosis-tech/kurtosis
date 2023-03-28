package feedback

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/savioxavier/termlink"
	"net/url"
	"strings"
)

const (
	commandShortDescription = "Give feedback"
	commandDescription      = "Give feedback, file a bug report or feature request, or get help from the Kurtosis team. " +
		"See below for the many ways you can get in touch with us.\n" +
		"TIP: You can quickly type and send us feedback directly from the CLI. For example, " +
		"kurtosis feedback \"I enjoy the enclave naming theme\" will open the Kurtosis Github " +
		"choose new issue type page with the description pre-filled with \"I enjoy the enclave naming theme\"."

	githubFlagKey                 = "github"
	emailFlagKey                  = "email"
	calendlyFlagKey               = "calendly"
	bugFeedbackFlagKey            = "bug"
	featureRequestFeedbackFlagKey = "feature"
	docsFeedbackFlagKey           = "docs"

	githubFlagUsageDescription    = "Takes you to our Github where you can file a bug report, feature request, or get help."
	emailFlagUsageDescription     = "Opens your mail client to send us feedback via email."
	calendlyFlagDescription       = "When set, opens the link to our Calendly page to schedule a 1:1 session with a Kurtosis expert."
	bugFlagDescription            = "To specify that this is a bug feedback type"
	featureRequestFlagDescription = "To specify that this is a feature request feedback type"
	docsFlagDescription           = "To specify that this is a docs feedback type"

	bugEmailSubjectPrefix       = "[BUG]"
	featureRequestSubjectPrefix = "[FEATURE_REQUEST]"
	docsSubjectPrefix           = "[DOCS]"

	defaultOpenGitHubIssuePage = "false"
	defaultOpenEmailLink       = "false"
	defaultOpenCalendlyLink    = "false"
	defaultBug                 = "false"
	defaultFeatureRequest      = "false"
	defaultDocs                = "false"

	githubLinkText     = "let us know in our Github."
	emailLinkText      = "click here to email us."
	onboardingLinkText = "schedule an on-boarding session with us."

	feedbackMsgTitle = "Your feedback is valuable and helps us improve Kurtosis. Thank you."

	feedbackMsg = `
This command can be used to deliver feedback to the Kurtosis team!

* Pass in your feedback as an argument using double quotations (e.g. kurtosis feedback "my feedback").
* Pass in the --github flag to open our Github Issues templates, pre-filled with your feedback/arg.
* Pass in the --email flag to open a draft email, pre-filled with your feedback/arg, to send to feedback@kurtosistech.com.
* Pass in the --calendly flag to open our Calendly link to schedule a 1:1 session with us for feedback and questions you may have!

See below for the some direct links as well.

* For bugs/issues, %v
* For general feedback, %v
* If you need help getting started, %v
`
	greenColorStr = "green"

	userMsgArgKey          = "message"
	userMsgArgDefaultValue = "default-message-value"
	userMsgArgIsOptional   = true
	userMsgArgIsNotGreedy  = false

	flagKeysStrSeparator = ", "

	notDestinationTypeSelected = ""
	notFeedbackTypeSelected    = ""
	emptyUserMsg               = ""

	defaultDestinationType = githubFlagKey
)

var FeedbackCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.FeedbackCmdStr,
	ShortDescription: commandShortDescription,
	LongDescription:  commandDescription,
	Flags: []*flags.FlagConfig{
		{
			Key:       githubFlagKey,
			Usage:     githubFlagUsageDescription,
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultOpenGitHubIssuePage,
		},
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
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultBug,
		},
		{
			Key:       featureRequestFeedbackFlagKey,
			Usage:     featureRequestFlagDescription,
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultFeatureRequest,
		},
		{
			Key:       docsFeedbackFlagKey,
			Usage:     docsFlagDescription,
			Shorthand: "",
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

	mutuallyExclusiveFeedbackDestinationTypeFlagKeys := []string{
		githubFlagKey, emailFlagKey, calendlyFlagKey}

	selectedDestinationType, err := validateMutuallyExclusiveBooleanFlagsAndGetSelectedKey(
		flags,
		mutuallyExclusiveFeedbackDestinationTypeFlagKeys,
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating mutually exclusive flags '%+v'",
			mutuallyExclusiveFeedbackDestinationTypeFlagKeys,
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

	//Github destination is used by default if users fill the message or any feedback type but do not specify the destination
	if selectedDestinationType == notDestinationTypeSelected &&
		(userMsg != emptyUserMsg || selectedFeedbackType != notFeedbackTypeSelected) {
		selectedDestinationType = defaultDestinationType
	}

	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics user id")
	}

	gitHubIssueURL, err := getGithubIssueURL(selectedFeedbackType, metricsUserId, userMsg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Github Issue URL")
	}

	if selectedDestinationType == githubFlagKey {
		if err := multi_os_command_executor.OpenFile(gitHubIssueURL); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Github issue page")
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

	//The following message is printed if destination is not set
	spotlightMessagePrinter := output_printers.GetSpotlightMessagePrinter()
	spotlightMessagePrinter.Print(feedbackMsgTitle)

	fmt.Printf(
		feedbackMsg,
		termlink.ColorLink(githubLinkText, gitHubIssueURL, greenColorStr),
		termlink.ColorLink(emailLinkText, user_support_constants.FeedbackEmailLink, greenColorStr),
		termlink.ColorLink(onboardingLinkText, user_support_constants.KurtosisOnBoardCalendlyUrl, greenColorStr),
	)
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
	userEncodedMsg := &url.URL{
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
		return stacktrace.NewError("Error validating the user message argument, it can be an empty string")
	}
	return nil
}

func getGithubIssueURL(
	selectedFeedbackType string,
	metricsUserId string,
	userEncodedMsgStr string,
) (string, error) {

	var gitHubIssueBaseUrl string
	switch selectedFeedbackType {
	case bugFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GithubBugIssueUrl
	case featureRequestFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GithubFeatureRequestIssueUrl
	case docsFeedbackFlagKey:
		gitHubIssueBaseUrl = user_support_constants.GithubDocsIssueUrl
	case notFeedbackTypeSelected:
		gitHubIssueBaseUrl = user_support_constants.GithubChooseNewIssuesUrl
	default:
		return "", stacktrace.NewError(
			"An error occurred while setting the Github URL, expected "+
				"to receive a know feedback type but '%v' was received instead; this is a bug in Kurtosis",
			selectedFeedbackType,
		)
	}

	gitHubIssueURL := fmt.Sprintf(
		"%s&version=%v&metrics-user-id=%v&description=%s&background-and-motivation=%s",
		gitHubIssueBaseUrl,
		kurtosis_version.KurtosisVersion,
		metricsUserId,
		userEncodedMsgStr,
		userEncodedMsgStr,
	)

	return gitHubIssueURL, nil
}

//TODO we could add this to the command framework
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

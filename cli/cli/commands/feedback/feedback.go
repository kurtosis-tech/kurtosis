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
)

const (
	commandShortDescription = "Give feedback"
	commandDescription      = "Give feedback, file a bug report or feature request, or get help from the Kurtosis team."

	githubFlagKey   = "github"
	emailFlagKey    = "email"
	calendlyFlagKey = "calendly"

	githubFlagUsageDescription = "Takes you to our Github where you can file a bug report, feature request, or get help."
	emailFlagUsageDescription  = "Opens your mail client to send us feedback via email."
	calendlyFlagDescription    = "When set, opens the link to our Calendly page to schedule a 1:1 session with a Kurtosis expert."

	defaultOpenGitHubIssuePage = "false"
	defaultOpenEmailLink       = "false"
	defaultOpenCalendlyLink    = "false"

	githubLinkText     = "let us know in our Github."
	emailLinkText      = "click here to email us."
	onboardingLinkText = "schedule an on-boarding session with us."

	feedbackMsgTitle = "Your feedback is valuable and helps us improve Kurtosis. Thank you."

	feedbackMsg = `
* For bugs/issues, %v
* For general feedback, %v
* If you need help getting started, %v
`
	greenColorStr = "green"

	userMsgArgKey          = "message"
	userMsgArgDefaultValue = "default-message-value"
	userMsgArgIsOptional   = true
	userMsgArgIsNotGreedy  = false
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
	isGithubFlagActivated := false
	isEmailFlagActivated := false
	isCalendlyFlagActivated := false
	doesUserFillMsg := true

	// Args parsing and validation
	userMsg, err := args.GetNonGreedyArg(userMsgArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the user message argument using flag key '%v'", userMsgArgKey)
	}
	if userMsg == userMsgArgDefaultValue {
		doesUserFillMsg = false
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

	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics user id")
	}

	isEmailFlagActivated, err = flags.GetBool(emailFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", emailFlagKey)
	}

	emailLink := fmt.Sprintf("%s?body=%s", user_support_constants.FeedbackEmailLink, userEncodedMsg)

	isGithubFlagActivated, err = flags.GetBool(githubFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", githubFlagKey)
	}

	gitHubIssueURL := fmt.Sprintf(
		"%s?version=%v&metrics-user-id=%v&description=%s&background-and-motivation=%s",
		user_support_constants.GithubIssuesUrl,
		kurtosis_version.KurtosisVersion,
		metricsUserId,
		userEncodedMsg,
		userEncodedMsg,
	)

	isCalendlyFlagActivated, err = flags.GetBool(calendlyFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", calendlyFlagKey)
	}

	if isEmailFlagActivated || (doesUserFillMsg && !isGithubFlagActivated) {
		if err := multi_os_command_executor.OpenFile(emailLink); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the feedback email link")
		}
		return nil
	}

	if isGithubFlagActivated {
		if err := multi_os_command_executor.OpenFile(gitHubIssueURL); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Github issue page")
		}
		return nil
	}

	if isCalendlyFlagActivated {
		if err := multi_os_command_executor.OpenFile(user_support_constants.KurtosisOnBoardCalendlyUrl); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Calendly email link")
		}
	}

	spotlightMessagePrinter := output_printers.GetSpotlightMessagePrinter()
	spotlightMessagePrinter.Print(feedbackMsgTitle)

	fmt.Printf(
		feedbackMsg,
		termlink.ColorLink(githubLinkText, gitHubIssueURL, greenColorStr),
		termlink.ColorLink(emailLinkText, emailLink, greenColorStr),
		termlink.ColorLink(onboardingLinkText, user_support_constants.KurtosisOnBoardCalendlyUrl, greenColorStr),
	)
	return nil
}

func validateUserMsgArg(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	userMsgArg, err := args.GetNonGreedyArg(userMsgArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the user message arguments using flag key '%v'", userMsgArgKey)
	}

	if userMsgArg == "" {
		return stacktrace.Propagate(err, "Error validating the user message argument, it can be an empty string")
	}
	return nil
}

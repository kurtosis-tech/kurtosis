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
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, flags *flags.ParsedFlags, _ *args.ParsedArgs) error {
	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics user id")
	}

	shouldOpenGitHubIssuesPage, err := flags.GetBool(githubFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", githubFlagKey)
	}

	gitHubIssueURL := fmt.Sprintf(
		"%s?version=%v&metrics-user-id=%v",
		user_support_constants.GithubIssuesUrl,
		kurtosis_version.KurtosisVersion,
		metricsUserId,
	)

	if shouldOpenGitHubIssuesPage {
		if err := multi_os_command_executor.OpenFile(gitHubIssueURL); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Github issue page")
		}
		return nil
	}

	shouldOpenEmailLink, err := flags.GetBool(emailFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", emailFlagKey)
	}

	body := fmt.Sprintf(
		"Thanks for sending us feedback, we value your thoughts.\n\n"+
			"We've prefilled some information about your Kurtosis CLI instance. "+
			"While it is optional to send this information to us, this metadata can be help our"+
			" team debug issues and turn around deliver fixes/improvements faster.\n"+
			"CLI version: %v\nMetrics User ID: %v", kurtosis_version.KurtosisVersion, metricsUserId)

	feedbackEmailLink := fmt.Sprintf(
		"%s?body=%v",
		user_support_constants.FeedbackEmailLink,
		body,
	)

	if shouldOpenEmailLink {
		if err := multi_os_command_executor.OpenFile(feedbackEmailLink); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the feedback email link")
		}
		return nil
	}

	shouldOpenCalendlyLink, err := flags.GetBool(calendlyFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", calendlyFlagKey)
	}

	if shouldOpenCalendlyLink {
		if err := multi_os_command_executor.OpenFile(user_support_constants.KurtosisOnBoardCalendlyUrl); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Calendly email link")
		}
	}

	spotlightMessagePrinter := output_printers.GetSpotlightMessagePrinter()
	spotlightMessagePrinter.Print(feedbackMsgTitle)

	fmt.Printf(
		feedbackMsg,
		termlink.ColorLink(githubLinkText, gitHubIssueURL, greenColorStr),
		termlink.ColorLink(emailLinkText, feedbackEmailLink, greenColorStr),
		termlink.ColorLink(onboardingLinkText, user_support_constants.KurtosisOnBoardCalendlyUrl, greenColorStr),
	)
	return nil
}

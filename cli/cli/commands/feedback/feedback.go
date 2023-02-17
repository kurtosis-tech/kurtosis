package feedback

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/savioxavier/termlink"
)

const (
	githubFlagKey = "github"
	emailFlagKey  = "email"

	defaultOpenGitHubIssuePage = "false"
	defaultOpenEmailLink       = "false"

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
	ShortDescription: "Give feedback",
	LongDescription:  "Give feedback, file a bug report or feature request, or get help from the Kurtosis team.",
	Flags: []*flags.FlagConfig{
		{
			Key:       githubFlagKey,
			Usage:     "Takes you to our Github where you can file a bug report, feature request, or get help.",
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultOpenGitHubIssuePage,
		},
		{
			Key:       emailFlagKey,
			Usage:     "Opens your mail client to send us feedback via email.",
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultOpenEmailLink,
		},
	},
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, flags *flags.ParsedFlags, _ *args.ParsedArgs) error {
	shouldOpenGitHubIssuesPage, err := flags.GetBool(githubFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", githubFlagKey)
	}

	if shouldOpenGitHubIssuesPage {
		if err := multi_os_command_executor.OpenFile(user_support_constants.GithubIssuesUrl); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis DiscordCmd Channel")
		}

		return nil
	}

	shouldOpenEmailLink, err := flags.GetBool(emailFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", emailFlagKey)
	}

	if shouldOpenEmailLink {
		if err := multi_os_command_executor.OpenFile(user_support_constants.FeedbackEmailLink); err != nil {
			return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis DiscordCmd Channel")
		}

		return nil
	}

	spotlightMessagePrinter := output_printers.GetSpotlightMessagePrinter()
	spotlightMessagePrinter.Print(feedbackMsgTitle)

	fmt.Printf(
		feedbackMsg,
		termlink.ColorLink(githubLinkText, user_support_constants.GithubIssuesUrl, greenColorStr),
		termlink.ColorLink(emailLinkText, user_support_constants.FeedbackEmailLink, greenColorStr),
		termlink.ColorLink(onboardingLinkText, user_support_constants.KurtosisOnBoardCalendlyUrl, greenColorStr),
	)
	return nil
}

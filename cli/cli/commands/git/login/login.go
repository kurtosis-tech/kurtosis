package login

import (
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"os"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorizes Kurtosis CLI on behalf of a git user",
	Long:  "Authorizes Kurtosis CLI's git operations",
	RunE:  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {

	secret, userLogin, err := AuthFlow("github.com", "", []string{}, true, *browser.New("", os.Stdout, os.Stderr))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred in the user login flow.")
	}
	logrus.Infof("Successfully authorized git user: %v", userLogin)

	//set password
	err = keyring.Set("kurtosis-git", userLogin, secret)
	if err != nil {
		logrus.Errorf("Unable to set token for keyring")
	}
	err = os.Setenv("GIT_USER", userLogin)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting git user env var.")
	}
	logrus.Debugf("Successfully set git token in keyring: %v", secret)
	logrus.Infof("Successfully set git auth info for user: %v", userLogin)
	return nil
}

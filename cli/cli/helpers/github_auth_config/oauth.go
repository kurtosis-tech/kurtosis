package github_auth_config

import (
	"bufio"
	"fmt"
	"github.com/cli/cli/v2/api"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/cli/oauth"
	"io"
	"net/http"
	"net/url"
	"os"
)

var (
	// The "Kurtosis CLI" OAuth app
	oauthClientID = "ff28fd26dcaf1be48c45"
	// This value is safe to be embedded in version control
	//TODO: verify that it's okay to embed client secret into version control
	oauthClientSecret = "***REMOVED***"

	githubHostname = "github.com"
)

func AuthFlow(oauthHost string, notice string, additionalScopes []string, isInteractive bool, b browser.Browser) (string, string, error) {
	httpClient := &http.Client{}

	minimumScopes := []string{"repo", "read:org", "gist"}
	scopes := append(minimumScopes, additionalScopes...)

	callbackURI := "http://127.0.0.1/callback"
	flow := &oauth.Flow{
		Host:         oauth.GitHubHost(githubHostname),
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		CallbackURI:  callbackURI,
		Scopes:       scopes,
		DisplayCode: func(code, verificationURL string) error {
			fmt.Fprintf(os.Stdout, "%s First copy your one-time code: %s\n", "!", code)
			return nil
		},
		BrowseURL: func(authURL string) error {
			if u, err := url.Parse(authURL); err == nil {
				if u.Scheme != "http" && u.Scheme != "https" {
					return fmt.Errorf("invalid URL: %s", authURL)
				}
			} else {
				return err
			}

			if !isInteractive {
				fmt.Fprintf(os.Stdout, "%s to continue in your web browser: %s\n", "Open this URL", authURL)
				return nil
			}

			fmt.Fprintf(os.Stdout, "%s to open %s in your browser... ", "Press Enter", oauthHost)
			_ = waitForEnter(os.Stdin)

			if err := b.Browse(authURL); err != nil {
				fmt.Fprintf(os.Stdout, "%s Failed opening a web browser at %s\n", "!", authURL)
				fmt.Fprintf(os.Stdout, "  %s\n", err)
				fmt.Fprint(os.Stdout, "  Please try entering the URL in your browser manually\n")
			}
			return nil
		},
		WriteSuccessHTML: func(w io.Writer) {
			fmt.Fprint(w, oauthSuccessPage)
		},
		HTTPClient: httpClient,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
	}

	fmt.Fprintln(os.Stdout, notice)

	token, err := flow.DetectFlow()
	if err != nil {
		return "", "", err
	}

	userLogin, err := getViewer(oauthHost, token.Token, os.Stderr)
	if err != nil {
		return "", "", err
	}

	return token.Token, userLogin, nil
}

type cfg struct {
	token string
}

func (c cfg) ActiveToken(hostname string) (string, string) {
	return c.token, "oauth_token"
}

func getViewer(hostname, token string, logWriter io.Writer) (string, error) {
	opts := api.HTTPClientOptions{
		Config: cfg{token: token},
		Log:    logWriter,
	}
	client, err := api.NewHTTPClient(opts)
	if err != nil {
		return "", err
	}
	return api.CurrentLoginName(api.NewClientFromHTTP(client), hostname)
}

func waitForEnter(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return scanner.Err()
}

const oauthSuccessPage = `
<!doctype html>
<meta charset="utf-8">
<title>Success: GitHub CLI</title>
<style type="text/css">
body {
  color: #1B1F23;
  background: #F6F8FA;
  font-size: 14px;
  font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif;
  line-height: 1.5;
  max-width: 620px;
  margin: 28px auto;
  text-align: center;
}

h1 {
  font-size: 24px;
  margin-bottom: 0;
}

p {
  margin-top: 0;
}

.box {
  border: 1px solid #E1E4E8;
  background: white;
  padding: 24px;
  margin: 28px;
}
</style>
<body>
  <svg height="52" class="octicon octicon-mark-github" viewBox="0 0 16 16" version="1.1" width="52" aria-hidden="true"><path fill-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path></svg>
  <div class="box">
    <h1>Successfully authenticated Kurtosis CLI</h1>
    <p>You may now close this tab and return to the terminal.</p>
  </div>
</body>
`

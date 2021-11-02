package version_checker

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	latestReleaseOnGitHubURL   = "https://api.github.com/repos/kurtosis-tech/kurtosis-cli-release-artifacts/releases/latest"
	acceptHttpHeaderKey        = "Accept"
	acceptHttpHeaderValue      = "application/json"
	contentTypeHttpHeaderKey   = "Content-Type"
	contentTypeHttpHeaderValue = "application/json"
	userAgentHttpHeaderKey     = "User-Agent"
	userAgentHttpHeaderValue   = "kurtosis-tech"
)

type GitHubReleaseReponse struct {
	TagName string `json:"tag_name"`
}

func CheckLatestVersion() {
	isLatestVersion, latestVersion, err := isLatestVersion()
	if err != nil {
		logrus.Warning("An error occurred trying to check if you are running the lates Kurtosis CLI version.")
		logrus.Debugf("Checking latest version error: '%v", err)
		logrus.Warningf("Your current version is '%v'", kurtosis_cli_version.KurtosisCLIVersion)
		logrus.Warningf("And you can manually check if your current version is the latest through this page: https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases")
		return
	}
	if !isLatestVersion {
		logrus.Warningf("You are running an old version of the Kurtosis CLI; we suggest you to update it to the latest version, '%v'", latestVersion)
		logrus.Warning("You can update it to the latest version doing the following steps")
		logrus.Warning("If you installed Kurtosis CLI with Brew, execute:")
		logrus.Warning("sudo brew uninstall kurtosis-tech/tap/kurtosis")
		logrus.Warning("sudo brew install kurtosis-tech/tap/kurtosis")
		logrus.Warning("================================================")
		logrus.Warning("If you installed Kurtosis CLI with APT, execute:")
		logrus.Warning("sudo apt install --only-upgrade kurtosis-cli")
		logrus.Warning("================================================")
		logrus.Warning("If you installed Kurtosis CLI with Yum, execute:")
		logrus.Warning("sudo yum upgrade kurtosis-cli")
		logrus.Warning("================================================")
		logrus.Warning("If you manually installed Kurtosis CLI with a DEB, RPM or an APL package:")
		logrus.Warning("Download the latest released package from the artifact page https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases")
	}
	return
}

func isLatestVersion() (bool, string, error) {
	ownVersion := kurtosis_cli_version.KurtosisCLIVersion
	latestVersion, err := getLatestReleaseVersionFromGitHub()
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred getting the latest release version number from the GitHub public API")
	}

	if ownVersion == latestVersion {
		return true, latestVersion, nil
	}

	return false, latestVersion, nil
}

func getLatestReleaseVersionFromGitHub() (string, error) {
	var (
		client         = &http.Client{}
		requestMethod  = "GET"
		requestBody    io.Reader
		responseObject GitHubReleaseReponse
	)

	request, err := http.NewRequest(requestMethod, latestReleaseOnGitHubURL, requestBody)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new HTTP GET request to URL '%v' ", latestReleaseOnGitHubURL)
	}

	request.Header.Add(acceptHttpHeaderKey, acceptHttpHeaderValue)
	request.Header.Add(contentTypeHttpHeaderKey, contentTypeHttpHeaderValue)
	request.Header.Add(userAgentHttpHeaderKey, userAgentHttpHeaderValue)

	response, err := client.Do(request)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing HTTP GET request to URL '%v' ", latestReleaseOnGitHubURL)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			logrus.Warnf("We tried to close the response body, but doing so threw an error:\n%v", err)
		}
	}()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the HTTP response body")
	}

	if err := json.Unmarshal(bodyBytes, &responseObject); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the latest release body response")
	}

	latestVersion := responseObject.TagName
	if latestVersion == "" {
		return "", stacktrace.Propagate(err, "The latest release version got from GitHub releases is empty")
	}

	return latestVersion, nil
}

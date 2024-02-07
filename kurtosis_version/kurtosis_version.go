package kurtosis_version

import "strings"

const Name = "Kurtosis"

var (
	AppName = ""
	Version = ""
	Commit  = ""
)

func GetAppName() string {
	return strings.TrimSpace(AppName)
}

func GetVersion() string {
	clean_kurtosis_version := strings.TrimSpace(Version)
	clean_commit := strings.TrimSpace(Commit)
	if clean_kurtosis_version != "" {
		return clean_kurtosis_version
	}
	if clean_commit != "" {
		return clean_commit
	}
	return "undefined"
}

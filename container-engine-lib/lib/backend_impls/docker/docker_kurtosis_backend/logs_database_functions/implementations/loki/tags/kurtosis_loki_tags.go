package docker_kurtosis_backend

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_labels_for_logs"
	"strings"
)

const (
	notAllowedCharsInLokiTags = " .-"
	noSeparationChar          = ""

	shouldChangeNextCharToUpperCaseInitialValue = false
)

// This is the list of all Logs Database Kurtosis Tracked Docker Labels
// converted in a valid Loki tags list.
// These tags are used for querying the logs database server in order to get
// the logs from an specific container
// Loki tags does not accept some characters like "."
func GetAllLogsDatabaseKurtosisTrackedValidLokiTags() []string {
	allLogsDatabaseKurtosisTrackedValidLokiTags := []string{}

	for _, logsDatabaseKurtosisTrackedDockerLabel := range docker_labels_for_logs.LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream {
		validLokiTag := newValidFormatLokiTagValue(logsDatabaseKurtosisTrackedDockerLabel.GetString())
		allLogsDatabaseKurtosisTrackedValidLokiTags = append(allLogsDatabaseKurtosisTrackedValidLokiTags, validLokiTag)
	}

	return allLogsDatabaseKurtosisTrackedValidLokiTags
}

func GetAllLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey() map[*docker_label_key.DockerLabelKey]string {
	allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey := map[*docker_label_key.DockerLabelKey]string{}

	for _, logsDatabaseKurtosisTrackedLabel := range docker_labels_for_logs.LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream {
		kurtosisTrackedValidLokiTag := newValidFormatLokiTagValue(logsDatabaseKurtosisTrackedLabel.GetString())
		allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey[logsDatabaseKurtosisTrackedLabel] = kurtosisTrackedValidLokiTag
	}

	return allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey
}

func newValidFormatLokiTagValue(stringToModify string) string {
	stringToModifyInLowerCase := strings.ToLower(stringToModify)
	shouldChangeNextCharToUpperCase := shouldChangeNextCharToUpperCaseInitialValue
	var shouldChangeCharToUpperCase bool
	var newString string
	for _, currenChar := range strings.Split(stringToModifyInLowerCase, noSeparationChar) {
		newChar := currenChar
		shouldChangeCharToUpperCase = shouldChangeNextCharToUpperCase
		if shouldChangeCharToUpperCase {
			newChar = strings.ToUpper(newChar)
		}
		if strings.ContainsAny(currenChar, notAllowedCharsInLokiTags) {
			shouldChangeNextCharToUpperCase = true
		} else {
			shouldChangeNextCharToUpperCase = false
			newString = newString + newChar
		}
	}
	return newString
}

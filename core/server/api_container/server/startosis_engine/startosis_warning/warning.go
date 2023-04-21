package startosis_warning

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

var warningMessageSet *sync.Map

// Initialize This method is responsible for creating warningMessageSet; if the warning set already exists it will override
// that with an empty warningMessageSet. This method is intended to be used during initialization process.
func Initialize() {
	warningMessageSet = new(sync.Map)
}

func PrintOnceAtTheEndOfExecutionf(message string, args ...interface{}) {
	if warningMessageSet != nil {
		formattedMessage := fmt.Sprintf(message, args...)
		_, stored := warningMessageSet.LoadOrStore(formattedMessage, true)
		if !stored {
			logrus.Warnf("Error occurred while adding warning to the set with message: %v", formattedMessage)
		}
	}
}

func GetContentFromWarningSet() []string {
	var warnings []string

	if warningMessageSet != nil {
		warningMessageSet.Range(func(key interface{}, value interface{}) bool {
			warnings = append(warnings, fmt.Sprint(key))
			return true
		})
	}

	return warnings
}

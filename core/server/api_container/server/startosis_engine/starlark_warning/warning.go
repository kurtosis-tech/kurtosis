package starlark_warning

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

/**
!!!! WARNING !!!!

THIS IS A GLOBAL VARIABLE THAT MAINTAINS THE STATE OF THE WARNING MESSAGES. THIS SET ENSURES THAT WE ONLY ONE WARNING PER
INSTRUCTION INVOCATION. WE KEEP TRACK OF ALL THE WARNINGS IN THIS SET, AND ONCE THE EXECUTION IS COMPLETED WE DISPLAY ALL
THE WARNINGS TO THE CLI, AND ALSO CLEAR THE SET - SO THAT FOR NEXT RUN WE HAVE AN EMPTY SET.

WE CURRENTLY DO IT BY HAVING A UNIQUE CONSTRAINT ON THE WARNING MESSAGE.THIS IS DONE JUST TO GET THE BALL ROLLING
- HOWEVER IN FUTURE - USING VISITOR PATTERN WHERE WE KEEP TRACK OF WHETHER ALL THE WARNINGS FROM AN INSTRUCTION/TYPE
IS RECEIVED - AND IF NOT THEN ADD IT OTHERWISE IGNORE.

MOST IMPORTANTLY - DO NOT DIRECTLY USE THIS VARIABLE AND INSTEAD USE PrintOnceAtTheEndOfExecutionf METHOD OR OTHER AVAILABLE
ABSTRACTIONS. THE LONG TERM PLAN IS EITHER TO EXTRACT THIS OUT INTO ITS OWN PACKAGE LIKE stacktrace OR PASS IT DOWN TO
INSTRUCTIONS ETC. I (PK) LEANING TOWARDS THE FIRST ONE, BECAUSE WARNING MESSAGES IS AN SEPARATE ENTITY AND DO NOT SEE ANY VALUE
ON COUPLING IT WITH OUR CORE LOGIC - THE WARNING SET ONLY CONTROLS THE STATE FOR WARNING MESSAGE WHICH CAN BE DONE IN ITS
OWN PACKAGE - HOWEVER IT IS NOT SUPER URGENT ATM
*/
var warningMessageSet *sync.Map

var once = new(sync.Once)

const warningMessageValue = true

// Clear the warning message set
// This method will force the methods to re-initialize the warning basically making the set empty
// This called everytime a startosis run is called
func Clear() {
	once = new(sync.Once)
}

// PrintOnceAtTheEndOfExecutionf This method stores the warnings in the warning set.
// The unique constraint is just the warning message, however,
//TODO: to have more comprehensive unique constraint such as instruction name
func PrintOnceAtTheEndOfExecutionf(message string, args ...interface{}) {
	once.Do(initialize)
	formattedMessage := fmt.Sprintf(message, args...)
	_, stored := warningMessageSet.LoadOrStore(formattedMessage, warningMessageValue)
	if !stored {
		logrus.Tracef("Error occurred while adding warning to the set with message: %v", formattedMessage)
	}
}

// GetContentFromWarningSet This method retrieves and deletes all the warnings from set. The main idea
// is that after calling this method the set will be empty and make sure that for next starlark run - we
// start with empty set
func GetContentFromWarningSet() []string {
	once.Do(initialize)
	var warnings []string
	warningMessageSet.Range(func(key interface{}, _ interface{}) bool {
		warnings = append(warnings, fmt.Sprint(key))
		// once the key is read, remove from the map
		warningMessageSet.Delete(key)
		return true
	})
	return warnings
}

// this is a private method and should not be called from anywhere outside this file
// using this method otherwise can result in inconsistent state and we won;t see warnings at all
func initialize() {
	warningMessageSet = new(sync.Map)
}

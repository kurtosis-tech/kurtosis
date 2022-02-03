package args

import (
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
	"strings"
)

const (
	jsonFieldTag          = "json"
)

// Fields are public for JSON de/serialization
type EngineServerArgs struct {
	ListenPortNum      uint16 `json:"listenPortNum"`

	LogLevelStr string	`json:"logLevelStr"`

	// So that the engine server knows its own version
	ImageVersionTag string `json:"imageVersionTag"`

	// The engine needs to know about this so it knows what filepath on the host machine to use when bind-mounting
	//  enclave data directories to the API container & services that the APIC starts
	EngineDataDirpathOnHostMachine string	`json:"engineDataDirpathOnHostMachine"`

	//The anonymized user ID for metrics analytics purpose `json:"metricsUserId"`
	MetricsUserID string `json:"metricsUserId"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`
}


// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewEngineServerArgs(
	listenPortNum uint16,
	logLevelStr string,
	imageVersionTag string,
	engineDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (*EngineServerArgs, error) {
	result := &EngineServerArgs{
		ListenPortNum:                  listenPortNum,
		LogLevelStr:                    logLevelStr,
		ImageVersionTag:                imageVersionTag,
		EngineDataDirpathOnHostMachine: engineDataDirpathOnHostMachine,
		MetricsUserID:                  metricsUserID,
		DidUserAcceptSendingMetrics:    didUserAcceptSendingMetrics,
	}

	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating engine server args")
	}
	return result, nil
}

func (args EngineServerArgs) validate() error {
	// Generic validation based on field type
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}

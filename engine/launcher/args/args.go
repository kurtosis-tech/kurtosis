package args

import (
	"reflect"
	"strings"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type EngineServerArgs struct {
	GrpcListenPortNum      uint16 `json:"grpcListenPortNum"`
	GrpcProxyListenPortNum uint16 `json:"grpcProxyListenPortNum"`

	LogLevelStr string `json:"logLevelStr"`

	// So that the engine server knows its own version
	ImageVersionTag string `json:"imageVersionTag"`

	// The engine needs to know about this so it knows what filepath on the host machine to use when bind-mounting
	//  enclave data directories to the API container & services that the APIC starts
	EngineDataDirpathOnHostMachine string `json:"engineDataDirpathOnHostMachine"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewEngineServerArgs(
	grpcListenPortNum uint16,
	grpcProxyListenPortNum uint16,
	logLevelStr string,
	imageVersionTag string,
	engineDataDirpathOnHostMachine string,
) (*EngineServerArgs, error) {
	result := &EngineServerArgs{
		GrpcListenPortNum:              grpcListenPortNum,
		GrpcProxyListenPortNum:         grpcProxyListenPortNum,
		LogLevelStr:                    logLevelStr,
		ImageVersionTag:                imageVersionTag,
		EngineDataDirpathOnHostMachine: engineDataDirpathOnHostMachine,
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
		field := reflectValType.Field(i)
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}

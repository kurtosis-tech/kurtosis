package vector

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

type VectorConfig struct {
	Sources map[string]map[string]interface{} `json:"sources,omitempty"`
	Sinks   map[string]map[string]interface{} `json:"sinks,omitempty"`
}

func newVectorConfig(
	listeningPortNumber uint16,
	sinks logs_aggregator.Sinks,
) *VectorConfig {
	reconciledSinks := map[string]map[string]interface{}{
		logs_aggregator.DefaultSinkId: {
			"type":   fileSinkType,
			"inputs": []string{defaultSourceId},
			"path":   uuidLogsFilepath,
			"encoding": map[string]interface{}{
				"codec": "json",
			},
			// Note: we set buffer to block so that we don't drop any logs, however this could apply backpressure up the topology
			// if we start noticing slowdown due to vector buffer blocking, we might want to revisit our architecture
			"buffer": map[string]interface{}{
				"when_full": "block",
			},
		},
	}

	for sinkId, sinkConfig := range sinks {
		reconciledSinks[sinkId] = map[string]interface{}{}
		for key, value := range sinkConfig {
			reconciledSinks[sinkId][key] = value
		}

		// Add inputs field to sink configuration
		reconciledSinks[sinkId]["inputs"] = []string{defaultSourceId}
	}

	return &VectorConfig{
		Sources: map[string]map[string]interface{}{
			defaultSourceId: {
				"type":    fluentBitSourceType,
				"address": fmt.Sprintf("%s:%s", fluentBitSourceIpAddress, strconv.Itoa(int(listeningPortNumber))),
			},
		},
		Sinks: reconciledSinks,
	}
}

func (cfg *VectorConfig) getConfigFileContent() (string, error) {
	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error marshalling config into JSON.")
	}

	return string(jsonBytes), nil
}

package vector

import (
	"fmt"
	"strconv"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

type VectorConfig struct {
	DataDir string                            `yaml:"data_dir"`
	Api     *VectorApiConfig                  `yaml:"api"`
	Sources map[string]map[string]interface{} `yaml:"sources,omitempty"`
	Sinks   map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

type VectorApiConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

func newVectorConfig(
	listeningPortNumber uint16,
	httpPortNumber uint16,
	sinks logs_aggregator.Sinks,
	shouldTurnOffPersistentVolumeLogsCollection bool,
) *VectorConfig {
	reconciledSinks := map[string]map[string]interface{}{}

	if !shouldTurnOffPersistentVolumeLogsCollection {
		reconciledSinks[logs_aggregator.DefaultSinkId] = map[string]interface{}{
			"type":   fileSinkType,
			"inputs": []string{defaultSourceId},
			"path":   uuidLogsFilepath,
			"encoding": map[string]interface{}{
				"codec": "json",
			},
			// Note: we set buffer to block so that we don't drop any logs, however this could apply backpressure up the topology
			// if we start noticing slowdown due to vector buffer blocking, we might want to revisit our architecture
			"buffer": map[string]interface{}{
				"type":      "disk",
				"max_size":  bufferSize,
				"when_full": "block",
			},
		}
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
		DataDir: vectorDataDirMountPath,
		Api: &VectorApiConfig{
			Enabled: true,
			Address: "0.0.0.0:" + strconv.Itoa(int(httpPortNumber)),
		},
		Sources: map[string]map[string]interface{}{
			defaultSourceId: {
				"type":    fluentBitSourceType,
				"address": fmt.Sprintf("%s:%s", fluentBitSourceIpAddress, strconv.Itoa(int(listeningPortNumber))),
			},
		},
		Sinks: reconciledSinks,
	}
}

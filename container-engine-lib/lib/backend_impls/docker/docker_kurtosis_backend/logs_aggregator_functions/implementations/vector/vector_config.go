package vector

import (
	"fmt"
	"strconv"
)

type VectorConfig struct {
	Source *Source `yaml:"sources"`
	Sink   *Sink   `yaml:"sinks"`
}

type Source struct {
	Id      string `yaml:"id"`
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
}

type Sink struct {
	Id     string   `yaml:"id"`
	Type   string   `yaml:"type"`
	Inputs []string `yaml:"inputs"`
}

func newDefaultVectorConfig(listeningPortNumber uint16) *VectorConfig {
	return &VectorConfig{
		Source: &Source{
			Id:      fluentBitSourceId,
			Type:    fluentBitSourceType,
			Address: fmt.Sprintf("%s:%s", fluentBitSourceIpAddress, strconv.Itoa(int(listeningPortNumber))),
		},
		Sink: &Sink{
			Id:     stdoutSinkID,
			Type:   stdoutTypeId,
			Inputs: []string{fluentBitSourceId},
		},
	}
}

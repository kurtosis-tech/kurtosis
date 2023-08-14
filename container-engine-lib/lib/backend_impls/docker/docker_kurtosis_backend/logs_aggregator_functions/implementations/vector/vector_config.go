package vector

import (
	"fmt"
	"strconv"
)

type VectorConfig struct {
	Source *Source
	Sink   *Sink
}

type Source struct {
	Id      string
	Type    string
	Address string
}

type Sink struct {
	Id     string
	Type   string
	Inputs []string
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

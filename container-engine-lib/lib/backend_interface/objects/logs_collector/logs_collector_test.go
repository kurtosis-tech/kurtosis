package logs_collector

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogsCollectorFiltersAgain(t *testing.T) {
	filter := Filter{
		Name:  "test",
		Match: "test",
		Params: []FilterParam{
			{
				Key:   "test",
				Value: "test",
			},
		},
	}

	jsonBytes, err := json.Marshal(filter)
	require.NoError(t, err)

	var unmarshalledFilter Filter
	err = json.Unmarshal(jsonBytes, &unmarshalledFilter)
	require.NoError(t, err)

	require.Equal(t, filter.Name, unmarshalledFilter.Name)
	require.Equal(t, filter.Match, unmarshalledFilter.Match)
	require.Equal(t, filter.Params, unmarshalledFilter.Params)
}

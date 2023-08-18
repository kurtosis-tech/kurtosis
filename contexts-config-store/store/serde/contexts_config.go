package serde

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/encoding/protojson"
)

func SerializeKurtosisContextsConfig(kurtosisContextsConfig *generated.KurtosisContextsConfig) ([]byte, error) {
	serializedKurtosisContext, err := protojson.Marshal(kurtosisContextsConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to serialize Kurtosis contexts config object")
	}
	return serializedKurtosisContext, nil
}

func DeserializeKurtosisContextsConfig(serializedKurtosisContextsConfig []byte) (*generated.KurtosisContextsConfig, error) {
	kurtosisContextsConfig := new(generated.KurtosisContextsConfig)
	unmarshaller := protojson.UnmarshalOptions{DiscardUnknown: true} // nolint: exhaustruct
	if err := unmarshaller.Unmarshal(serializedKurtosisContextsConfig, kurtosisContextsConfig); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to deserialize Kurtosis contexts config object")
	}
	return kurtosisContextsConfig, nil
}

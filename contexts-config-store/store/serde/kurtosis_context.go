package serde

import (
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/encoding/protojson"
)

func SerializeKurtosisContext(kurtosisContext *generated.KurtosisContext) ([]byte, error) {
	serializedKurtosisContext, err := protojson.Marshal(kurtosisContext)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to serialize Kurtosis context object")
	}
	return serializedKurtosisContext, nil
}

func DeserializeKurtosisContext(serializedKurtosisContext []byte) (*generated.KurtosisContext, error) {
	kurtosisContext := new(generated.KurtosisContext)
	if err := protojson.Unmarshal(serializedKurtosisContext, kurtosisContext); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to deserialize Kurtosis context object")
	}
	return kurtosisContext, nil
}

package to_grpc

import (
	"fmt"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

func ToGrpcConnect(conn api_type.Connect) rpc_api.Connect {
	switch conn {
	case api_type.CONNECT:
		return rpc_api.Connect_CONNECT
	case api_type.NOCONNECT:
		return rpc_api.Connect_NO_CONNECT
	default:
		panic(fmt.Sprintf("Missing conversion of Connect Enum value: %s", conn))
	}
}

func ToGrpcFeatureFlag(flag api_type.KurtosisFeatureFlag) rpc_api.KurtosisFeatureFlag {
	switch flag {
	case api_type.NOINSTRUCTIONSCACHING:
		return rpc_api.KurtosisFeatureFlag_NO_INSTRUCTIONS_CACHING
	default:
		panic(fmt.Sprintf("Missing conversion of Feature Flag Enum value: %s", flag))
	}
}

func ToGrpcImageDownloadMode(flag api_type.ImageDownloadMode) rpc_api.ImageDownloadMode {
	switch flag {
	case api_type.ImageDownloadModeALWAYS:
		return rpc_api.ImageDownloadMode_always
	case api_type.ImageDownloadModeMISSING:
		return rpc_api.ImageDownloadMode_missing
	default:
		panic(fmt.Sprintf("Missing conversion of Image Download Mode Enum value: %s", flag))
	}
}

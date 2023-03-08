package recipe

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

const (
	emptyServiceName = service.ServiceName("")
)

func MakeOptional(argName string) string {
	return fmt.Sprintf("%s?", argName)
}

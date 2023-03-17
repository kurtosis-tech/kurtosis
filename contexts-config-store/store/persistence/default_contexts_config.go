package persistence

import (
	"github.com/google/uuid"
	api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	defaultContextName = "default"
)

func NewDefaultContextsConfig() (*generated.KurtosisContextsConfig, error) {
	// This logic is similar to uuid_generate.GenerateUUIDString, but we don't want to pull container engine lib here
	// just for this.
	// TODO: extract the UUID generator outside of container engine lib maybe
	randomUuid, err := uuid.NewRandom()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to generate a random UUID for the default context")
	}
	newContextUuidStr := strings.Replace(randomUuid.String(), "-", "", -1)

	defaultContextUuid := api.NewContextUuid(newContextUuidStr)
	defaultContext := api.NewLocalOnlyContext(defaultContextUuid, defaultContextName)
	return api.NewKurtosisContextsConfig(defaultContextUuid, defaultContext), nil
}

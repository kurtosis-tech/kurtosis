package persistence

import api "github.com/kurtosis-tech/kurtosis/contexts-state-store/api/golang"

var defaultContextName = "default"
var defaultContextUuid = api.NewContextUuid("00000000000000000000000000000000")

var defaultContext = api.NewLocalOnlyContext(defaultContextUuid, defaultContextName)

var defaultContextsState = api.NewKurtosisContextsState(defaultContextUuid, defaultContext)

package persistence

import api "github.com/kurtosis-tech/kurtosis/context-config-store/api/golang"

var personalContextName = "personal"
var personnalContextUuid = api.NewContextUuid("00000000-0000-0000-0000-000000000000")

var personalContext = api.NewLocalOnlyContext(personnalContextUuid, personalContextName)

var defaultContextConfig = api.NewKurtosisContextConfig(personnalContextUuid, personalContext)

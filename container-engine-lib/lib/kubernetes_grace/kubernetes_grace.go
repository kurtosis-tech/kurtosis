package kubernetes_grace

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// EnvVarName overrides the grace period for deleting Kurtosis' pods (and the
// TerminationGracePeriodSeconds stamped on their specs). Unset keeps Kubernetes'
// default; "0" deletes immediately for fast teardown of ephemeral pods.
const EnvVarName = "KURTOSIS_POD_DELETE_GRACE_PERIOD_SECONDS"

// Override returns the configured grace period, or nil to keep Kubernetes' default.
func Override() *int64 {
	raw := os.Getenv(EnvVarName)
	if raw == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		logrus.Warnf("Ignoring invalid %v=%q; expected a non-negative integer", EnvVarName, raw)
		return nil
	}
	return &parsed
}

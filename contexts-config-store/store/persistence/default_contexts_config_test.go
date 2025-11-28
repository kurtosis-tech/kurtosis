package persistence

import (
	api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	// Duplicate default context name value on purpose here to fail loudly if someone change it in the other file by
	// mistake
	expectedDefaultContextName = "default"
)

func TestNewDefaultContextsConfig(t *testing.T) {
	defaultContextsConfig, err := NewDefaultContextsConfig()
	require.NoError(t, err)

	require.Len(t, defaultContextsConfig.GetContexts(), 1)

	onlyContext := defaultContextsConfig.GetContexts()[0]

	// validate context UUID looks ok and matches the current context UUID
	defaultContextUuidStr := onlyContext.GetUuid().GetValue()
	require.Regexp(t, "[a-z0-9]{32}", defaultContextUuidStr)
	require.Equal(t, defaultContextUuidStr, defaultContextsConfig.GetCurrentContextUuid().GetValue())

	// validate context is local and name is correct.
	require.Equal(t, onlyContext.GetName(), expectedDefaultContextName)
	_, err = api.Visit[struct{}](onlyContext, api.KurtosisContextVisitor[struct{}]{
		VisitRemoteContextV0: func(remoteContext *generated.RemoteContextV0) (*struct{}, error) {
			return nil, stacktrace.NewError("default context should be a local-only context!")
		},
		VisitLocalOnlyContextV0: func(localContext *generated.LocalOnlyContextV0) (*struct{}, error) {
			return nil, nil
		},
	})
	require.NoError(t, err)
}

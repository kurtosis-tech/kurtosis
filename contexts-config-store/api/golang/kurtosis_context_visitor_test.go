package golang

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	contextUuid = NewContextUuid("context-uuid")
	contextName = "context-name"

	visitorTestResult = "pong"
)

func TestLocalOnlyContext(t *testing.T) {
	kurtosisContext := NewLocalOnlyContext(contextUuid, contextName)

	result, err := Visit[string](kurtosisContext, KurtosisContextVisitor[string]{
		VisitLocalOnlyContextV0: func(localOnlyContext *generated.LocalOnlyContextV0) (*string, error) {
			return &visitorTestResult, nil
		},
		VisitRemoteContextV0: func(remoteContext *generated.RemoteContextV0) (*string, error) {
			return nil, stacktrace.NewError("Should not be called")
		},
	})
	require.Nil(t, err)
	require.Equal(t, visitorTestResult, *result)
}

func TestRemoteContext(t *testing.T) {
	kurtosisContext := &generated.KurtosisContext{
		Uuid: contextUuid,
		Name: contextName,
		KurtosisContextInfo: &generated.KurtosisContext_RemoteContextV0{
			// Disabling exhaustruct here as building an entire RemoteContextV0 just for this test seems unnecessary
			// nolint: exhaustruct
			RemoteContextV0: &generated.RemoteContextV0{},
		},
	}

	result, err := Visit[string](kurtosisContext, KurtosisContextVisitor[string]{
		VisitLocalOnlyContextV0: func(localOnlyContext *generated.LocalOnlyContextV0) (*string, error) {
			return nil, stacktrace.NewError("Should not be called")
		},
		VisitRemoteContextV0: func(remoteContext *generated.RemoteContextV0) (*string, error) {
			return &visitorTestResult, nil
		},
	})
	require.Nil(t, err)
	require.Equal(t, visitorTestResult, *result)
}

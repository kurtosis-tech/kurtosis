package golang

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var (
	contextUuid = &ContextUuid{Value: "context-uuid"}
	contextName = "context-name"

	visitorTestResult = "pong"
)

func TestVisitorCoversAllFields(t *testing.T) {
	tmp := new(isKurtosisContext_KurtosisContextInfo)
	elem := reflect.ValueOf(tmp).Elem()
	typ := reflect.TypeOf(*tmp)
	fmt.Println(typ)
	fmt.Println(elem)
}

func TestLocalOnlyContext(t *testing.T) {
	kurtosisContext := &KurtosisContext{
		Uuid: contextUuid,
		Name: contextName,
		KurtosisContextInfo: &KurtosisContext_LocalOnlyContextV0{
			LocalOnlyContextV0: &LocalOnlyContextV0{},
		},
	}

	result, err := Visit[string](kurtosisContext, KurtosisContextVisitor[string]{
		VisitLocalOnlyContextV0: func(localOnlyContext *LocalOnlyContextV0) (*string, error) {
			return &visitorTestResult, nil
		},
		VisitRemoteContextV0: func(remoteContext *RemoteContextV0) (*string, error) {
			return nil, stacktrace.NewError("Should not be called")
		},
	})
	require.Nil(t, err)
	require.Equal(t, visitorTestResult, *result)
}

func TestRemoteContext(t *testing.T) {
	kurtosisContext := &KurtosisContext{
		Uuid: contextUuid,
		Name: contextName,
		KurtosisContextInfo: &KurtosisContext_RemoteContextV0{
			RemoteContextV0: &RemoteContextV0{},
		},
	}

	result, err := Visit[string](kurtosisContext, KurtosisContextVisitor[string]{
		VisitLocalOnlyContextV0: func(localOnlyContext *LocalOnlyContextV0) (*string, error) {
			return nil, stacktrace.NewError("Should not be called")
		},
		VisitRemoteContextV0: func(remoteContext *RemoteContextV0) (*string, error) {
			return &visitorTestResult, nil
		},
	})
	require.Nil(t, err)
	require.Equal(t, visitorTestResult, *result)
}

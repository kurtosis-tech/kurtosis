package golang

import (
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
)

type KurtosisContextVisitor[ResultType any] struct {
	VisitLocalOnlyContextV0 func(localContext *LocalOnlyContextV0) (*ResultType, error)

	VisitRemoteContextV0 func(localContext *RemoteContextV0) (*ResultType, error)
}

func Visit[ResultType any](kurtosisContext *KurtosisContext, visitor KurtosisContextVisitor[ResultType]) (*ResultType, error) {
	if kurtosisContext.GetLocalOnlyContextV0() != nil {
		return visitor.VisitLocalOnlyContextV0(kurtosisContext.GetLocalOnlyContextV0())
	} else if kurtosisContext.GetRemoteContextV0() != nil {
		return visitor.VisitRemoteContextV0(kurtosisContext.GetRemoteContextV0())
	}
	return nil, stacktrace.NewError("Type of KurtosisContext couldn't be resolved: '%s'", reflect.TypeOf(kurtosisContext.KurtosisContextInfo))
}

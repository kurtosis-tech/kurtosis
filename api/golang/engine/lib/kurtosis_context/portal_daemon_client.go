package kurtosis_context

import (
	"context"
	"fmt"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	contexts_store_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	contexts_store_generated_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DefaultGrpcPortalClientPortNum = uint16(9731)
)

// CreatePortalDaemonClient builds a portal daemon GRPC client based on the current context and a
// forLocalContextReturnNilIfUnreachable flag
// If the flag is set to true, it returns a nil client for local contexts when the portal is not reachable.
// Local context are working seamlessly without the portal (unless the user wants to map port). Therefore we're not
// requiring it for most workflows
// TODO: when it's assumed that Kurtosis Portal should be running everywhere, we can remove this logic and just build a
// regular client
func CreatePortalDaemonClient(currentContext *contexts_store_generated_api.KurtosisContext, forLocalContextReturnNilIfUnreachable bool) (portal_api.KurtosisPortalClientClient, error) {
	// When the context is remote, we build a client to the locally running portal daemon
	kurtosisPortalSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, DefaultGrpcPortalClientPortNum)
	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	portalConn, err := grpc.Dial(kurtosisPortalSocketStr, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Portal Client at '%v'",
			kurtosisPortalSocketStr,
		)
	}
	portalClient := portal_api.NewKurtosisPortalClientClient(portalConn)
	_, portalReachableError := portalClient.Ping(context.Background(), portal_constructors.NewPortalPing())

	visitLocalContext := func(localContext *contexts_store_generated_api.LocalOnlyContextV0) (*struct{}, error) {
		if portalReachableError != nil && forLocalContextReturnNilIfUnreachable {
			logrus.Infof("Portal daemon unreachable on port '%d' but will not fail as the context in use is a local context: '%s'", DefaultGrpcPortalClientPortNum, currentContext.GetName())
			// error is allowed here, overriding the portalClient to nil as we can't connect to it but this is
			// more or less expected based on the forLocalContextReturnNilIfUnreachable flag
			portalClient = nil
			return nil, nil
		}
		if portalReachableError != nil {
			return nil, portalReachableError
		}
		return nil, nil
	}
	visitRemoteContext := func(remoteContext *contexts_store_generated_api.RemoteContextV0) (*struct{}, error) {
		if portalReachableError != nil {
			return nil, portalReachableError
		}
		return nil, nil
	}
	instantiatePortalClientVisitor := contexts_store_api.KurtosisContextVisitor[struct{}]{
		VisitLocalOnlyContextV0: visitLocalContext,
		VisitRemoteContextV0:    visitRemoteContext,
	}
	_, err = contexts_store_api.Visit[struct{}](currentContext, instantiatePortalClientVisitor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building client for Kurtosis Portal daemon")
	}
	return portalClient, nil
}

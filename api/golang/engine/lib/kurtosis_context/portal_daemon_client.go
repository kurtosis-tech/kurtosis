package kurtosis_context

import (
	"context"
	"fmt"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DefaultGrpcPortalClientPortNum = uint16(9731)
)

// CreatePortalDaemonClient builds a portal daemon GRPC client based on the current context and a
// forLocalContextReturnNilIfUnreachable flag
// If the flag is set to true, it returns an error if the Portal cannot be reached. If false, it returns a nil client.
// This is necessary as Portal is not required. If/When it is, this flag can be removed
func CreatePortalDaemonClient(mustBuildClient bool) (portal_api.KurtosisPortalClientClient, error) {
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
	if portalReachableError != nil {
		if mustBuildClient {
			return nil, stacktrace.Propagate(portalReachableError, "Kurtosis Portal unreachable")
		}
		logrus.Debugf("Kurtosis Portal daemon is currently not reachable. If Kurtosis is being used on" +
			"a local-only context, this is fine as Portal is not required for local-only contexts.")
		// not error-ing here since Portal is optional for now
		return nil, nil
	}
	return portalClient, nil
}

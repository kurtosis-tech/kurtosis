package kurtosis_context

import (
	"context"
	"fmt"
	"time"

	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultGrpcPortalClientPortNum = uint16(9731)

	waitForPortalClientPingTimeout = 5 * time.Second
)

// CreatePortalDaemonClient builds a portal daemon GRPC client.
func CreatePortalDaemonClient() (portal_api.KurtosisPortalClientClient, error) {
	kurtosisPortalSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, DefaultGrpcPortalClientPortNum)
	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	portalConn, err := grpc.Dial(kurtosisPortalSocketStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Portal Client at '%v'",
			kurtosisPortalSocketStr,
		)
	}
	portalClient := portal_api.NewKurtosisPortalClientClient(portalConn)
	ctxWithTimeout, cancelFunc := context.WithTimeout(context.Background(), waitForPortalClientPingTimeout)
	defer cancelFunc()
	_, portalReachableError := portalClient.Ping(ctxWithTimeout, portal_constructors.NewPortalPing(), grpc.WaitForReady(true))
	if portalReachableError != nil {
		return nil, stacktrace.Propagate(portalReachableError, "Kurtosis Portal unreachable")
	}
	return portalClient, nil
}

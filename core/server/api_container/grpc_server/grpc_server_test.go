package grpc_server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	testToken           = "test-token"
	testStopGracePeriod = 2 * time.Second
	testCallTimeout     = 5 * time.Second
)

func getFreePort(t *testing.T) uint16 {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return uint16(port)
}

func startAuthenticatedHealthServer(t *testing.T, token string) grpc_health_v1.HealthClient {
	port := getFreePort(t)
	server := NewGrpcServer(port, testStopGracePeriod, auth.NewGrpcServerOptions(token), []func(*grpc.Server){
		func(grpcServer *grpc.Server) {
			grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
		},
	})
	stopper := make(chan struct{}, 1)
	serverDone := make(chan error, 1)
	go func() { serverDone <- server.RunUntilStopped(stopper) }()
	t.Cleanup(func() {
		stopper <- struct{}{}
		require.NoError(t, <-serverDone)
	})

	conn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%v", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return grpc_health_v1.NewHealthClient(conn)
}

func checkHealth(client grpc_health_v1.HealthClient, authorizationHeader string) error {
	ctx, cancel := context.WithTimeout(context.Background(), testCallTimeout)
	defer cancel()
	if authorizationHeader != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authorizationHeader)
	}
	_, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: ""}, grpc.WaitForReady(true))
	return err
}

func TestServerRejectsCallsWithoutTheToken(t *testing.T) {
	client := startAuthenticatedHealthServer(t, testToken)

	require.NoError(t, checkHealth(client, "Bearer "+testToken))
	require.Equal(t, codes.Unauthenticated, status.Code(checkHealth(client, "")))
	require.Equal(t, codes.Unauthenticated, status.Code(checkHealth(client, "Bearer wrong")))
}

func TestServerWithoutATokenRejectsEverything(t *testing.T) {
	client := startAuthenticatedHealthServer(t, "")

	require.Equal(t, codes.Unauthenticated, status.Code(checkHealth(client, "")))
	require.Equal(t, codes.Unauthenticated, status.Code(checkHealth(client, "Bearer ")))
}

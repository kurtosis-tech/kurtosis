package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testToken = "test-token"

func serveWithMiddleware(token string, method string, authorizationHeader string) int {
	handler := NewHttpMiddleware(token)(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(method, "/", nil)
	if authorizationHeader != "" {
		request.Header.Set(httpHeaderName, authorizationHeader)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestHttpMiddlewareRejectsMissingOrWrongToken(t *testing.T) {
	require.Equal(t, http.StatusUnauthorized, serveWithMiddleware(testToken, http.MethodPost, ""))
	require.Equal(t, http.StatusUnauthorized, serveWithMiddleware(testToken, http.MethodPost, "Bearer wrong"))
	require.Equal(t, http.StatusUnauthorized, serveWithMiddleware(testToken, http.MethodPost, testToken))
	require.Equal(t, http.StatusOK, serveWithMiddleware(testToken, http.MethodPost, "Bearer "+testToken))
}

func TestHttpMiddlewareLetsPreflightThroughAndIsDisabledWithoutToken(t *testing.T) {
	require.Equal(t, http.StatusOK, serveWithMiddleware(testToken, http.MethodOptions, ""))
	require.Equal(t, http.StatusOK, serveWithMiddleware("", http.MethodPost, ""))
}

func TestGrpcContextCheck(t *testing.T) {
	require.Equal(t, codes.Unauthenticated, status.Code(checkGrpcContext(context.Background(), testToken)))

	wrongCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcMetadataKey, "Bearer wrong"))
	require.Equal(t, codes.Unauthenticated, status.Code(checkGrpcContext(wrongCtx, testToken)))

	validCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcMetadataKey, "Bearer "+testToken))
	require.NoError(t, checkGrpcContext(validCtx, testToken))
}

func TestTokenIsReadFromEnvVarAndCreatedOnce(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(EnvVarName, "")
	xdg.Reload()
	t.Cleanup(xdg.Reload)

	token, err := GetToken()
	require.NoError(t, err)
	require.Empty(t, token)

	createdToken, err := GetOrCreateToken()
	require.NoError(t, err)
	require.Len(t, createdToken, tokenNumBytes*2)

	sameToken, err := GetOrCreateToken()
	require.NoError(t, err)
	require.Equal(t, createdToken, sameToken)

	t.Setenv(EnvVarName, "from-env")
	envToken, err := GetToken()
	require.NoError(t, err)
	require.Equal(t, "from-env", envToken)
}

func serveWithHostValidation(host string, domain string) int {
	handler := NewHostValidationMiddleware(domain)(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = host
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestHostValidationRejectsForeignHostnames(t *testing.T) {
	for _, allowedHost := range []string{"localhost:9710", "127.0.0.1:9779", "[::1]:9711", "172.17.0.3:9710", "engine", "LOCALHOST:8081"} {
		require.Equal(t, http.StatusOK, serveWithHostValidation(allowedHost, ""), allowedHost)
	}
	for _, forbiddenHost := range []string{"evil.example", "evil.example:9779", "localhost.evil.example:9711", ""} {
		require.Equal(t, http.StatusForbidden, serveWithHostValidation(forbiddenHost, ""), forbiddenHost)
	}
	require.Equal(t, http.StatusOK, serveWithHostValidation("kurtosis.example.com", "kurtosis.example.com"))
}

func TestAllowedBrowserOriginsAreTheWebUIOnly(t *testing.T) {
	require.Equal(t, []string{"http://localhost:9711", "http://127.0.0.1:9711"}, GetAllowedBrowserOrigins(""))
	require.Contains(t, GetAllowedBrowserOrigins("kurtosis.example.com"), "https://kurtosis.example.com")
	require.NotContains(t, GetAllowedBrowserOrigins(""), "*")
}

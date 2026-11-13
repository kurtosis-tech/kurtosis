package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	EnvVarName = "KURTOSIS_API_AUTH_TOKEN"

	applicationDirname = "kurtosis"
	tokenFileName      = "api-auth-token"
	tokenFilePerm      = 0600
	tokenNumBytes      = 32

	httpHeaderName       = "Authorization"
	grpcMetadataKey      = "authorization"
	bearerPrefix         = "Bearer "
	unauthorizedMessage  = "missing or invalid Kurtosis API auth token"
	forbiddenHostMessage = "request Host is not allowed"

	contentTypeHeaderName     = "Content-Type"
	grpcContentTypePrefix     = "application/grpc"
	grpcStatusHeaderName      = "Grpc-Status"
	grpcMessageHeaderName     = "Grpc-Message"
	grpcUnauthenticatedStatus = "16"

	localhostHostname      = "localhost"
	engineProxyHostname    = "engine"
	webUIPort              = "9711"
	TokenUrlFragmentPrefix = "kurtosis-token="
)

func GetToken() (string, error) {
	if token := strings.TrimSpace(os.Getenv(EnvVarName)); token != "" {
		return token, nil
	}
	tokenFilePath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}
	tokenBytes, err := os.ReadFile(tokenFilePath)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the Kurtosis API auth token file '%v'", tokenFilePath)
	}
	return strings.TrimSpace(string(tokenBytes)), nil
}

func GetOrCreateToken() (string, error) {
	token, err := GetToken()
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
	tokenBytes := make([]byte, tokenNumBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the Kurtosis API auth token")
	}
	token = hex.EncodeToString(tokenBytes)
	tokenFilePath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(tokenFilePath, []byte(token), tokenFilePerm); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing the Kurtosis API auth token file '%v'", tokenFilePath)
	}
	return token, nil
}

func GetServerToken() (string, error) {
	token := strings.TrimSpace(os.Getenv(EnvVarName))
	if token == "" {
		return "", stacktrace.NewError("Refusing to serve the Kurtosis API without an auth token; set the '%v' environment variable", EnvVarName)
	}
	return token, nil
}

func GrpcDialOption() grpc.DialOption {
	return grpc.WithPerRPCCredentials(tokenCredentials{})
}

func NewHttpClient(token string) *http.Client {
	// nolint: exhaustruct
	return &http.Client{Transport: tokenRoundTripper{token: token}}
}

func NewHttpMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.Method == http.MethodOptions || isValidBearer(request.Header.Get(httpHeaderName), token) {
				next.ServeHTTP(writer, request)
				return
			}
			writeUnauthorized(writer, request)
		})
	}
}

func writeUnauthorized(writer http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get(contentTypeHeaderName)
	if strings.HasPrefix(contentType, grpcContentTypePrefix) {
		writer.Header().Set(contentTypeHeaderName, contentType)
		writer.Header().Set(grpcStatusHeaderName, grpcUnauthenticatedStatus)
		writer.Header().Set(grpcMessageHeaderName, unauthorizedMessage)
		writer.WriteHeader(http.StatusOK)
		return
	}
	http.Error(writer, unauthorizedMessage, http.StatusUnauthorized)
}

func NewHostValidationMiddleware(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !isAllowedHost(request.Host, domain) {
				http.Error(writer, forbiddenHostMessage, http.StatusForbidden)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func NewApiMiddleware(token string, domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewHostValidationMiddleware(domain)(NewHttpMiddleware(token)(next))
	}
}

func GetAllowedBrowserOrigins(domain string) []string {
	origins := []string{
		"http://" + net.JoinHostPort(localhostHostname, webUIPort),
		"http://" + net.JoinHostPort("127.0.0.1", webUIPort),
	}
	if domain != "" {
		origins = append(origins, "https://"+domain, "http://"+domain)
	}
	return origins
}

func isAllowedHost(hostHeader string, domain string) bool {
	hostname := hostHeader
	if splitHostname, _, err := net.SplitHostPort(hostHeader); err == nil {
		hostname = splitHostname
	}
	hostname = strings.ToLower(strings.Trim(hostname, "[]"))
	if net.ParseIP(hostname) != nil {
		return true
	}
	if hostname == localhostHostname || hostname == engineProxyHostname {
		return true
	}
	return domain != "" && hostname == strings.ToLower(domain)
}

func NewGrpcServerOptions(token string) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if err := checkGrpcContext(ctx, token); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}),
		grpc.ChainStreamInterceptor(func(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if err := checkGrpcContext(stream.Context(), token); err != nil {
				return err
			}
			return handler(srv, stream)
		}),
	}
}

func checkGrpcContext(ctx context.Context, token string) error {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	for _, headerValue := range requestMetadata.Get(grpcMetadataKey) {
		if isValidBearer(headerValue, token) {
			return nil
		}
	}
	return status.Error(codes.Unauthenticated, unauthorizedMessage)
}

func isValidBearer(headerValue string, token string) bool {
	if token == "" {
		return false
	}
	presentedToken, found := strings.CutPrefix(headerValue, bearerPrefix)
	return found && subtle.ConstantTimeCompare([]byte(presentedToken), []byte(token)) == 1
}

func getTokenFilePath() (string, error) {
	tokenFilePath, err := xdg.ConfigFile(path.Join(applicationDirname, tokenFileName))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis API auth token file path")
	}
	return tokenFilePath, nil
}

type tokenCredentials struct{}

func (creds tokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}
	if token == "" {
		return map[string]string{}, nil
	}
	return map[string]string{grpcMetadataKey: bearerPrefix + token}, nil
}

func (creds tokenCredentials) RequireTransportSecurity() bool {
	return false
}

var _ credentials.PerRPCCredentials = tokenCredentials{}

type tokenRoundTripper struct {
	token string
}

func (roundTripper tokenRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if roundTripper.token == "" {
		return http.DefaultTransport.RoundTrip(request)
	}
	authedRequest := request.Clone(request.Context())
	authedRequest.Header.Set(httpHeaderName, bearerPrefix+roundTripper.token)
	return http.DefaultTransport.RoundTrip(authedRequest)
}

package upload_files

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsHTTPURL(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"https URL", "https://gist.githubusercontent.com/fjl/abc/raw/def/dashboard.json", true},
		{"http URL", "http://example.com/file.txt", true},
		{"github package locator (no scheme)", "github.com/foo/bar/file.star", false},
		{"absolute path", "/etc/hosts", false},
		{"relative path", "./dashboard.json", false},
		{"empty string", "", false},
		{"scheme but no host", "https://", false},
		{"unsupported scheme", "ftp://example.com/file.txt", false},
		{"file scheme", "file:///tmp/x.json", false},
		{"uppercase scheme is fine", "HTTPS://example.com/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isHTTPURL(tt.src))
		})
	}
}

func TestDownloadFileFromURL_Success(t *testing.T) {
	const body = "hello dashboard"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/raw/dashboard.json", r.URL.Path)
		_, err := w.Write([]byte(body))
		require.NoError(t, err)
	}))
	defer server.Close()

	downloadedPath, tempDir, err := downloadFileFromURL(context.Background(), server.URL+"/raw/dashboard.json")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.Equal(t, "dashboard.json", filepath.Base(downloadedPath))

	got, err := os.ReadFile(downloadedPath)
	require.NoError(t, err)
	require.Equal(t, body, string(got))
}

func TestDownloadFileFromURL_FallbackFilename(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	defer server.Close()

	// Trailing slash means path.Base returns "/", which should trigger the fallback name.
	downloadedPath, tempDir, err := downloadFileFromURL(context.Background(), server.URL+"/")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.Equal(t, defaultDownloadedFileName, filepath.Base(downloadedPath))
}

func TestDownloadFileFromURL_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer server.Close()

	_, tempDir, err := downloadFileFromURL(context.Background(), server.URL+"/missing.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")
	// On error the temp directory must not leak.
	_, statErr := os.Stat(tempDir)
	require.True(t, os.IsNotExist(statErr), "expected temp dir to be cleaned up on error, got stat err = %v", statErr)
}

func TestDescribeURLSource(t *testing.T) {
	require.Equal(t,
		"https://example.com/dashboard.json",
		describeURLSource("https://example.com/dashboard.json?token=secret#frag"),
	)
	require.Equal(t, "not a url", describeURLSource("not a url"))
}

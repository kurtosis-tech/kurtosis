package upload_files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	httpScheme  = "http"
	httpsScheme = "https"

	urlDownloadDirPattern = "kurtosis-upload-files-url-*"

	urlDownloadHTTPTimeout = 5 * time.Minute

	defaultDownloadedFileName = "download"
)

// isHTTPURL reports whether src is an absolute http or https URL.
// Package locators such as `github.com/foo/bar` have no scheme and therefore
// are not matched here; they continue to flow through the package content
// provider as before.
func isHTTPURL(src string) bool {
	parsed, err := url.Parse(src)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != httpScheme && scheme != httpsScheme {
		return false
	}
	return parsed.Host != ""
}

// downloadFileFromURL fetches src over HTTP(S) and writes its body to a file in
// a freshly created temporary directory. The directory path is returned so the
// caller can remove it once the contents have been compressed.
//
// The returned file path preserves the URL basename when one is present so the
// resulting files artifact uses an intuitive filename (matching the behaviour
// of upload_files for local sources).
func downloadFileFromURL(ctx context.Context, src string) (downloadedFilePath string, tempDir string, err error) {
	parsedURL, err := url.Parse(src)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to parse '%s' as a URL", src)
	}

	tempDir, err = os.MkdirTemp("", urlDownloadDirPattern)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to create temporary directory for URL download '%s'", src)
	}
	cleanupOnError := true
	defer func() {
		if cleanupOnError {
			_ = os.RemoveAll(tempDir)
		}
	}()

	fileName := path.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = defaultDownloadedFileName
	}
	downloadedFilePath = filepath.Join(tempDir, fileName)

	httpCtx, cancel := context.WithTimeout(ctx, urlDownloadHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, http.MethodGet, src, nil)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to build HTTP request for '%s'", src)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to download '%s'", src)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", stacktrace.NewError("Downloading '%s' returned unexpected HTTP status '%s'", src, resp.Status)
	}

	out, err := os.Create(downloadedFilePath)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to create temp file '%s' for URL download", downloadedFilePath)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		return "", "", stacktrace.Propagate(err, "Failed to write response body for '%s' to '%s'", src, downloadedFilePath)
	}
	if err := out.Close(); err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to close downloaded file '%s'", downloadedFilePath)
	}

	cleanupOnError = false
	return downloadedFilePath, tempDir, nil
}

// describeURLSource returns a short, deterministic string used as the artifact
// description default and YAML-plan src field when the source is a URL. It
// hides query strings and fragments to keep descriptions stable.
func describeURLSource(src string) string {
	parsed, err := url.Parse(src)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return src
	}
	return fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, parsed.Path)
}

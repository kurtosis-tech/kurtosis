package kurtosis_instruction

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	starlarkPath            = "/starlark/"
	packagePerm             = 0755
	temporaryRepoDirPattern = "tmp-repo-dir-*"
)

type CacheEntry struct {
	globals starlark.StringDict
	err     error
}

type PackageManager struct {
	enclaveDataVolume string
	cache             map[string]*CacheEntry
	packageTmpDir     string
	packageDir        string
}

func NewPackageManager(enclaveDataVolumeDirPath string) *PackageManager {
	return &PackageManager{
		enclaveDataVolume: enclaveDataVolumeDirPath,
		cache:             make(map[string]*CacheEntry),
		packageTmpDir:     path.Join(enclaveDataVolumeDirPath, "tmp"),
		packageDir:        path.Join(enclaveDataVolumeDirPath, starlarkPath),
	}
}

func (p *PackageManager) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	e, ok := p.cache[module]
	if e == nil {
		if ok {
			// request for package whose loading is in progress
			return nil, fmt.Errorf("cycle in load graph")
		}

		// Add a placeholder to indicate "load in progress".
		p.cache[module] = nil

		// Load it.
		contents, err := p.getPackageData(module)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while fetching contents of the module '%v'", module)
		}

		thread := &starlark.Thread{Name: "exec " + module, Load: thread.Load}
		globals, err := starlark.ExecFile(thread, module, contents, nil)
		e = &CacheEntry{globals, err}

		// Update the cache.
		p.cache[module] = e
	}
	return e.globals, e.err
}

func (p *PackageManager) getPackageData(githubURL string) (string, error) {
	parsedUrl, err := url.Parse(githubURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error parsing the url '%v'", githubURL)
	}
	if parsedUrl.Scheme != "https" {
		return "", stacktrace.NewError("Expected the scheme to be 'https' got '%v'", parsedUrl.Scheme)
	}
	if parsedUrl.Host != "github.com" {
		return "", stacktrace.NewError("We only support packages on Github for now")
	}

	splitURLPath := removeEmpty(strings.Split(parsedUrl.Path, "/"))

	if len(splitURLPath) < 2 {
		return "", stacktrace.NewError("URL path should contain at least 2 parts")
	}

	contents, err := os.ReadFile(p.getPathToStartosisFile(splitURLPath))
	if err == nil {
		return string(contents), nil
	}

	firstTwoSubPaths := strings.Join(splitURLPath[:2], "/")
	authorName := splitURLPath[0]
	gitRepo := "https://github.com/" + firstTwoSubPaths

	tempRepoDirPath, err := os.MkdirTemp(p.packageTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error creating temporary directory for repo")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, firstTwoSubPaths)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: gitRepo, Progress: io.Discard})
	if err != nil {
		return "", stacktrace.Propagate(err, "Error in cloning git repo '%v'", gitRepo)
	}
	packageAuthorPath := path.Join(p.packageDir, authorName)
	packagePath := path.Join(p.packageDir, firstTwoSubPaths)
	_, err = os.Stat(packageAuthorPath)
	if err != nil {
		if err = os.Mkdir(packageAuthorPath, packagePerm); err != nil {
			stacktrace.Propagate(err, "An error occurred while creating directory '%v'", packageAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, packagePath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while moving cloned file at temporary destination to final destination")
	}

	contents, err = os.ReadFile(p.getPathToStartosisFile(splitURLPath))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the StarLark file")
	}

	return string(contents), nil
}

func (p *PackageManager) getPathToStartosisFile(splitUrlPath []string) string {
	lastItem := splitUrlPath[len(splitUrlPath)-1]
	if !strings.HasSuffix(lastItem, ".star") {
		if len(splitUrlPath) > 2 {
			splitUrlPath[len(splitUrlPath)-1] = splitUrlPath[len(splitUrlPath)-1] + ".star"
		} else {
			splitUrlPath = append(splitUrlPath, "main.star")
		}
	}
	splitUrlPath = append([]string{p.packageDir}, splitUrlPath...)
	filePath := path.Join(splitUrlPath...)
	return filePath
}

func removeEmpty(splitPath []string) []string {
	var splitWithoutEmpties []string
	for _, subPath := range splitPath {
		if subPath != "" {
			splitWithoutEmpties = append(splitWithoutEmpties, subPath)
		}
	}
	return splitWithoutEmpties
}
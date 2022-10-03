package module_manager

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	starlarkSubpath         = "starlark"
	tmpSubpath              = "tmp"
	packageDirPermission    = 0755
	temporaryRepoDirPattern = "tmp-repo-dir-*"
)

type GitModuleManager struct {
	moduleTmpDir string
	moduleDir    string
}

func NewGitModuleManager(enclaveDataVolumeDir string) (*GitModuleManager, error) {
	moduleTmpDir := path.Join(enclaveDataVolumeDir, tmpSubpath)
	if _, err := os.Stat(moduleTmpDir); err != nil {
		if err = os.Mkdir(moduleTmpDir, packageDirPermission); err != nil {
			panic(fmt.Sprintf("Error creating temporary directory for repositories to be cloned '%v'", moduleTmpDir))
		}
	}
	moduleDir := path.Join(enclaveDataVolumeDir, starlarkSubpath)
	if _, err := os.Stat(moduleDir); err != nil {
		if err = os.Mkdir(moduleDir, packageDirPermission); err != nil {
			panic(fmt.Sprintf("Error creating directory for modules '%v'", moduleDir))
		}
	}

	return &GitModuleManager{
		moduleTmpDir: moduleTmpDir,
		moduleDir:    moduleDir,
	}, nil
}

func (p *GitModuleManager) GetModule(githubURL string) (string, error) {
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

	tempRepoDirPath, err := os.MkdirTemp(p.moduleTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, firstTwoSubPaths)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: gitRepo, Progress: io.Discard})
	if err != nil {
		return "", stacktrace.Propagate(err, "Error in cloning git repository '%v'", gitRepo)
	}
	packageAuthorPath := path.Join(p.moduleDir, authorName)
	packagePath := path.Join(p.moduleDir, firstTwoSubPaths)
	_, err = os.Stat(packageAuthorPath)
	if err != nil {
		if err = os.Mkdir(packageAuthorPath, packageDirPermission); err != nil {
			stacktrace.Propagate(err, "An error occurred while creating directory '%v'", packageAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, packagePath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while moving module at temporary destination to final destination")
	}

	contents, err = os.ReadFile(p.getPathToStartosisFile(splitURLPath))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the StarLark file")
	}

	return string(contents), nil
}

func (p *GitModuleManager) getPathToStartosisFile(splitUrlPath []string) string {
	lastItem := splitUrlPath[len(splitUrlPath)-1]
	if !strings.HasSuffix(lastItem, ".star") {
		if len(splitUrlPath) > 2 {
			splitUrlPath[len(splitUrlPath)-1] = splitUrlPath[len(splitUrlPath)-1] + ".star"
		} else {
			splitUrlPath = append(splitUrlPath, "main.star")
		}
	}
	splitUrlPath = append([]string{p.moduleDir}, splitUrlPath...)
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

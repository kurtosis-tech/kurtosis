package kurtosis_instruction

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"net/url"
)

type CacheEntry struct {
	globals starlark.StringDict
	err     error
}

func MakeLoad() func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	var cache = make(map[string]*CacheEntry)

	return func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		e, ok := cache[module]
		if e == nil {
			if ok {
				// request for package whose loading is in progress
				return nil, fmt.Errorf("cycle in load graph")
			}

			// Add a placeholder to indicate "load in progress".
			cache[module] = nil

			// Load it.
			thread := &starlark.Thread{Name: "exec " + module, Load: thread.Load}
			globals, err := starlark.ExecFile(thread, module, nil, nil)
			e = &CacheEntry{globals, err}

			// Update the cache.
			cache[module] = e
		}
		return e.globals, e.err
	}
}

func getPackageData(githubURL string) (string, error) {
	parsedUrl, err := url.Parse(githubURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error parsing the url '%v'", githubURL)
	}
	if parsedUrl.Scheme != "https" {
		return "", stacktrace.Propagate(err, "Expected the scheme to be 'https' got '%v'", parsedUrl.Scheme)
	}
	if parsedUrl.Host != "github.com" {
		return "", stacktrace.Propagate(err, "We only support packages on Github for now")
	}

	parsedUrl.Pa

	return "", nil
}

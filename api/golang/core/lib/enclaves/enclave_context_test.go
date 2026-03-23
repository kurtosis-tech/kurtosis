/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package enclaves

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsLocalDependencyReplace(t *testing.T) {
	tests := []struct {
		name     string
		replace  string
		expected bool
	}{
		{
			name:     "absolute path",
			replace:  "/tmp/my-package",
			expected: true,
		},
		{
			name:     "relative path with dot prefix",
			replace:  "./my-package",
			expected: true,
		},
		{
			name:     "relative path with double dot prefix",
			replace:  "../my-package",
			expected: true,
		},
		{
			name:     "remote github package",
			replace:  "github.com/kurtosis-tech/postgres-package",
			expected: false,
		},
		{
			name:     "empty string",
			replace:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalDependencyReplace(tt.replace)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalPackagePathResolution(t *testing.T) {
	packageRootPath := "/tmp/pkg2"

	tests := []struct {
		name          string
		replaceOption string
		expectedPath  string
	}{
		{
			name:          "absolute path is not joined with package root",
			replaceOption: "/tmp/pkg1",
			expectedPath:  "/tmp/pkg1",
		},
		{
			name:          "relative path is joined with package root",
			replaceOption: "../pkg1",
			expectedPath:  "/tmp/pkg1",
		},
		{
			name:          "relative dot path is joined with package root",
			replaceOption: "./nested/pkg1",
			expectedPath:  "/tmp/pkg2/nested/pkg1",
		},
		{
			name:          "deeply nested absolute path is preserved",
			replaceOption: "/home/user/projects/my-package",
			expectedPath:  "/home/user/projects/my-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localPackagePath := tt.replaceOption
			if !path.IsAbs(localPackagePath) {
				localPackagePath = path.Join(packageRootPath, localPackagePath)
			}
			require.Equal(t, tt.expectedPath, localPackagePath)
		})
	}
}

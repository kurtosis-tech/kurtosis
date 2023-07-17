package inspect

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"testing"
)

const expectedTreeStr = `artifact
└── path
    ├── to
    │   ├── file.txt [2.0K]
    │   └── another.txt [ 123]
    └── yet_another.txt [2.3M]
`

func TestTreeBuilding(t *testing.T) {
	treeStr := buildTree("artifact", []*kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription{
		{
			Path: "path/to/file.txt",
			Size: 2000,
		},
		{
			Path: "path/to/another.txt",
			Size: 123,
		},
		{
			Path: "path/yet_another.txt",
			Size: 2*1024*1024 + 300*1024,
		},
	})
	require.Equal(t, treeStr, expectedTreeStr)
}

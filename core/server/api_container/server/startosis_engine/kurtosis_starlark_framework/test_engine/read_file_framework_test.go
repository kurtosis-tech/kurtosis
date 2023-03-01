package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type readFileTestCase struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func newReadFileTestCase(t *testing.T) *readFileTestCase {
	packageContentProvider := startosis_packages.NewMockPackageContentProvider(t)
	packageContentProvider.EXPECT().GetModuleContents(TestModuleFileName).Return("Hello World!", nil)
	return &readFileTestCase{
		T:                      t,
		packageContentProvider: packageContentProvider,
	}
}

func (t *readFileTestCase) GetId() string {
	return read_file.ReadFileBuiltinName
}

func (t *readFileTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	return read_file.NewReadFileHelper(t.packageContentProvider)
}

func (t *readFileTestCase) GetStarlarkCode() string {
	return fmt.Sprintf(`%s(%s=%q)`, read_file.ReadFileBuiltinName, read_file.SrcArgName, TestModuleFileName)
}

func (t *readFileTestCase) Assert(result starlark.Value) {
	t.packageContentProvider.AssertCalled(t, "GetModuleContents", TestModuleFileName)
	require.Equal(t, result, starlark.String("Hello World!"))
}

package render_templates_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	enclaveTestName       = "render-templates-test"
	isPartitioningEnabled = false

	rootFilename      = "config.yml"
	nestedRelFilepath = "grafana/config.yml"
	expectedContents  = "Hello Stranger. The sum of [1 2 3] is 6. My favorite moment in history 1257894000. My favorite number 1231231243.43."

	fileServerServiceId services.ServiceID = "file-server"
)

func TestRenderTemplates(t *testing.T) {
	ctx := context.Background()
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	templateAndDataByDestRelFilepath := getTemplateAndDataByDestRelFilepath()

	filesArtifactUUID, err := enclaveCtx.RenderTemplates(templateAndDataByDestRelFilepath)
	require.NoError(t, err)

	fileServerPublicIp, fileServerPublicPortNum, err := test_helpers.StartFileServer(fileServerServiceId, filesArtifactUUID, rootFilename, enclaveCtx)
	require.NoError(t, err)

	err = testRenderedTemplates(templateAndDataByDestRelFilepath, fileServerPublicIp, fileServerPublicPortNum)
	require.NoError(t, err)
}

// ========================================================================
// Helpers
// ========================================================================

// Checks templates are rendered correctly and to the right files in the right subdirectories
func testRenderedTemplates(
	templateDataByDestinationFilepath map[string]*enclaves.TemplateAndData,
	ipAddress string,
	portNum uint16,
) error {

	for renderedTemplateFilepath := range templateDataByDestinationFilepath {
		if err := test_helpers.CheckFileContents(ipAddress, portNum, renderedTemplateFilepath, expectedContents); err != nil {
			return stacktrace.Propagate(err, "There was an error testing the content of file '%s'.", renderedTemplateFilepath)
		}
	}
	return nil
}

func getTemplateAndDataByDestRelFilepath() map[string]*enclaves.TemplateAndData {
	templateAndDataByDestRelFilepath := make(map[string]*enclaves.TemplateAndData)

	template := "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
	templateData := map[string]interface{}{"Name": "Stranger", "Answer": 6, "Numbers": []int{1, 2, 3}, "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43}
	templateAndData := enclaves.NewTemplateAndData(template, templateData)

	templateAndDataByDestRelFilepath[nestedRelFilepath] = templateAndData
	templateAndDataByDestRelFilepath[rootFilename] = templateAndData

	return templateAndDataByDestRelFilepath
}

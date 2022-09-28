import {err, ok, Result} from "neverthrow";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {checkFileContents, startFileServer} from "../../test_helpers/test_helpers";
import {TemplateAndData} from "kurtosis-sdk/build/core/lib/enclaves/template_and_data";
import {ServiceID} from "kurtosis-sdk";

const ENCLAVE_TEST_NAME         = "render-templates-test"
const IS_PARTITIONING_ENABLED   = false

const ROOT_FILENAME         = "config.yml"
const NESTED_REL_FILEPATH       = "grafana/config.yml"
const EXPECTED_CONTENTS = "Hello Stranger. The sum of [1 2 3] is 6. My favorite moment in history 1257894000. My favorite number 1231231243.43."

const FILE_SERVER_SERVICE_ID : ServiceID = "file-server"

jest.setTimeout(180000)

test("Test Render Templates", TestRenderTemplates)

async function TestRenderTemplates() {
    const createEnclaveResult = await createEnclave(ENCLAVE_TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value
    try {
        const templateAndDataByDestRelFilepath = getTemplateAndDataByDestRelFilepath()
        const renderTemplatesResults = await enclaveContext.renderTemplates(templateAndDataByDestRelFilepath)
        if(renderTemplatesResults.isErr()) { throw renderTemplatesResults.error }

        const filesArtifactUuid = renderTemplatesResults.value

        const startFileServerResult = await startFileServer(FILE_SERVER_SERVICE_ID, filesArtifactUuid, ROOT_FILENAME, enclaveContext)
        if (startFileServerResult.isErr()){throw startFileServerResult.error}
        const {fileServerPublicIp, fileServerPublicPortNum} = startFileServerResult.value

        const testRenderedTemplatesResult = await testRenderedTemplates(templateAndDataByDestRelFilepath, fileServerPublicIp, fileServerPublicPortNum)
        if(testRenderedTemplatesResult.isErr()) { throw testRenderedTemplatesResult.error}
    } finally {
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
}

//========================================================================
// Helpers
//========================================================================

// Checks templates are rendered correctly and to the right files in the right subdirectories
async function testRenderedTemplates(
    templateDataByDestinationFilepath : Map<string, TemplateAndData>,
    ipAddress: string,
    portNum: number,
): Promise<Result<null, Error>> {

    for (let [renderedTemplateFilepath, _] of templateDataByDestinationFilepath) {
        let testContentResults = await checkFileContents(ipAddress, portNum, renderedTemplateFilepath, EXPECTED_CONTENTS)
        if (testContentResults.isErr()) { return  err(testContentResults.error) }
    }
    return ok(null)
}

function getTemplateAndDataByDestRelFilepath() : Map<string, TemplateAndData> {
    let templateDataByDestinationRelFilepath = new Map<string, TemplateAndData>()

    const template = "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
    const templateData  = {"Name": "Stranger", "Answer": 6, "Numbers": [1, 2, 3], "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43}
    const templateAndData = new TemplateAndData(template, templateData)

    templateDataByDestinationRelFilepath.set(NESTED_REL_FILEPATH, templateAndData)
    templateDataByDestinationRelFilepath.set(ROOT_FILENAME, templateAndData)

    return templateDataByDestinationRelFilepath
}

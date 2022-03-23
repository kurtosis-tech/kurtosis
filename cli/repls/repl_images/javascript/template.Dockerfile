# IMPORTANT: must match the Kurt Client Javascript SDK version
FROM node:16.13.0-alpine

WORKDIR {{ .PackageInstallationDirpath }}

RUN npm install kurtosis-core-api-lib@{{ .KurtosisCoreVersion }}

WORKDIR /repl

ENV NODE_PATH="{{ .InstalledPackagesDirpath }}"

# Even though async/await is enabled for the REPL, for some reason the code ran with "-e" can't use it so we have to use
#  the old callback syntax to load Kurtosis (not a big deal though)
CMD node -i --experimental-repl-await -e " \
    let kurtosisCore = require(\"kurtosis-core-api-lib\"); \
    let enclaveCtx; \
    kurtosisCore.EnclaveContext.newGrpcNodeEnclaveContext( \
        \"${{ "{" }}{{ .KurtosisAPIContainerIPEnvVar }}{{ "}" }}\", \
        ${{ "{" }}{{ .KurtosisAPIContainerPortEnvVar }}{{ "}" }}, \
        \"${{ "{" }}{{ .EnclaveIDEnvVar }}{{ "}" }}\", \
        \"${{ "{" }}{{ .EnclaveDataMountDirpathEnvVar }}{{ "}" }}\" \
    ).then(newEnclaveCtxResp => { \
        if (newEnclaveCtxResp.isErr()) { \
            console.log(newEnclaveCtxResp.error); \
            process.exit(1); \
        } \
        enclaveCtx = newEnclaveCtxResp.value; \
    }); \
"

# IMPORTANT: must match the Kurt Client Javascript SDK version
FROM node:16.7.0-alpine

WORKDIR {{ .PackageInstallationDirpath }}

RUN npm install kurtosis-core-api-lib@{{ .KurtosisClientVersion }}

WORKDIR /repl

ENV NODE_PATH="{{ .InstalledPackagesDirpath }}"

# Even though async/await is enabled for the REPL, for some reason the code ran with "-e" can't use it so we have to use
#  the old callback syntax to load Kurtosis (not a big deal though)
CMD node -i --experimental-repl-await -e " \
    let kurtosisCore = require(\"kurtosis-core-api-lib\"); \
    let grpc = require(\"grpc\"); \
    let networkCtx; \
    const client = new kurtosisCore.ApiContainerServiceClient(\"${KURTOSIS_API_SOCKET}\", grpc.credentials.createInsecure()); \
    networkCtx = new kurtosisCore.NetworkContext(client, \"${ENCLAVE_DATA_VOLUME_MOUNTPOINT}\"); \
"

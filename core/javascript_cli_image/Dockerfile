# TODO must match the Kurt Client Javascript SDK version
FROM node:16.7.0-alpine

WORKDIR /repl

# TODO Make this auto-update somehow??? Or add a CI check to make sure that it matches the dependencies used in the internal testsuite??
RUN npm install kurtosis-core-api-lib@0.16.0

CMD node -i --experimental-repl-await -e " \
    let kurtosis; \
    let networkCtx; \
    import(\"kurtosis-core-api-lib\").then(kurtosisCore => { \
        kurtosis = kurtosisCore; \
        import(\"grpc\").then(grpc => { \
            const client = new kurtosisCore.ApiContainerServiceClient(\"${KURTOSIS_API_SOCKET}\", grpc.credentials.createInsecure()); \
            networkCtx = new kurtosisCore.NetworkContext(client, \"${ENCLAVE_DATA_VOLUME_MOUNTPOINT}\"); \
        }); \
    }); \
"

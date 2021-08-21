# TODO must match the Kurt Client Javascript SDK version
FROM node:16.7.0-alpine

WORKDIR /repl

RUN npm install kurtosis-core-api-lib@0.12.3

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

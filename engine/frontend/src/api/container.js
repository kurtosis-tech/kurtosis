import {
    getServicesFromEnclaveManager,
    inspectFilesArtifactContentsFromEnclaveManager,
    listFilesArtifactNamesAndUuidsFromEnclaveManager,
    runStarlarkPackageFromEnclaveManager
} from "./api";

const TransportProtocolEnum = ["tcp", "sctp", "udp"];

export const runStarlarkPackage = async (host, port, packageId, args, token) => {
    return runStarlarkPackageFromEnclaveManager(host, port, packageId, args, token)

    // const containerClient = new ApiContainerServicePromiseClient(url);
    // const runStarlarkPackageArgs = new RunStarlarkPackageArgs();
    //
    // runStarlarkPackageArgs.setDryRun(false);
    // runStarlarkPackageArgs.setRemote(true);
    // runStarlarkPackageArgs.setPackageId(packageId);
    // runStarlarkPackageArgs.setSerializedParams(args)
    // const stream = containerClient.runStarlarkPackage(runStarlarkPackageArgs, null);
    // return stream;x
}

const getDataFromApiContainer = async (request, process, token) => {
    const data = await request(token)
    return process(data)
}

export const getFileArtifactInfo = async (host, port, fileArtifactName, token) => {
    const makeGetFileArtifactInfo = async () => {
        try {
            return inspectFilesArtifactContentsFromEnclaveManager(host, port, fileArtifactName, token)
        } catch (error) {
            return {
                files: []
            }
        }
    }

    const processFileArtifact = (data) => {
        let processed = {}
        const recursive = (sub_dirs, index, processed) => {
            if (sub_dirs.length > index) {
                const subDir = sub_dirs[index]
                if (!processed[subDir] && subDir !== "") {
                    processed[subDir] = {}
                }
                recursive(sub_dirs, index + 1, processed[subDir])
            }
        }

        const add_file_recursive = (sub_dirs, index, processed, file) => {
            const subDir = sub_dirs[index]
            if (index === sub_dirs.length - 1) {
                processed[subDir] = {...processed[subDir], ...file}
            } else {
                add_file_recursive(sub_dirs, index + 1, processed[subDir], file)
            }
        }

        data.fileDescriptions.map(file => {
            const splitted_path = file.path.split("/")
            if (splitted_path[splitted_path.length - 1] === "") {
                recursive(splitted_path, 0, processed)
            } else {
                add_file_recursive(splitted_path, 0, processed, file)
            }
        })

        return {
            files: processed
        }
    }

    const response = await getDataFromApiContainer(makeGetFileArtifactInfo, processFileArtifact, token)
    return response;
}

export const getEnclaveInformation = async (host, port, token) => {
    if (host === "") {
        return {
            services: [],
            artifacts: []
        }
    }

    const makeGetServiceRequest = async () => {
        try {
            return await getServicesFromEnclaveManager(host, port, token);
        } catch (error) {
            return {serviceInfo: []}
        }
    }

    const makeFileArtifactRequest = async () => {
        return listFilesArtifactNamesAndUuidsFromEnclaveManager(host, port, token);
    }

    const processServiceRequest = (data) => {
        return Object.keys(data.serviceInfo)
            .map(serviceName => {
                const ports = Object.keys(data.serviceInfo[serviceName].maybePublicPorts)
                    .map(portName => {
                            return {
                                publicPortNumber: data.serviceInfo[serviceName].maybePublicPorts[portName].number,
                                privatePortNumber: data.serviceInfo[serviceName].privatePorts[portName].number,
                                applicationProtocol: data.serviceInfo[serviceName].privatePorts[portName].maybeApplicationProtocol,
                                portName: portName,
                                transportProtocol: TransportProtocolEnum[data.serviceInfo[serviceName].privatePorts[portName].transportProtocol],
                            }
                        }
                    )

                return {
                    name: data.serviceInfo[serviceName].name,
                    uuid: data.serviceInfo[serviceName].serviceUuid,
                    privateIpAddr: data.serviceInfo[serviceName].privateIpAddr,
                    ports: ports,
                }
            })
    }

    const processFileArtifactRequest = (data) => {
        return data.fileNamesAndUuids.map(artifact => {
            return {
                name: artifact.fileName,
                uuid: artifact.fileUuid,
            }
        })
    }

    const servicesPromise = getDataFromApiContainer(makeGetServiceRequest, processServiceRequest)
    const fileArtifactsPromise = getDataFromApiContainer(makeFileArtifactRequest, processFileArtifactRequest)

    const [services, artifacts] = await Promise.all([servicesPromise, fileArtifactsPromise])
    // console.log("sa", services, artifacts)
    return {services, artifacts}
}

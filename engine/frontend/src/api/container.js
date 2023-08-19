import google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb.js'
import {
    FilesArtifactNameAndUuid,
    InspectFilesArtifactContentsRequest,
    RunStarlarkPackageArgs
} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import {
    ApiContainerServicePromiseClient
} from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb'
import {getServicesFromEnclaveManager} from "./api";

const TransportProtocolEnum = ["tcp", "sctp", "udp"];

export const runStarlarkPackage = async (url, packageId, args) => {
    const containerClient = new ApiContainerServicePromiseClient(url);
    const runStarlarkPackageArgs = new RunStarlarkPackageArgs();

    runStarlarkPackageArgs.setDryRun(false);
    runStarlarkPackageArgs.setRemote(true);
    runStarlarkPackageArgs.setPackageId(packageId);
    runStarlarkPackageArgs.setSerializedParams(args)
    const stream = containerClient.runStarlarkPackage(runStarlarkPackageArgs, null);
    return stream;
}

const getDataFromApiContainer = async (request, process) => {
    const data = await request()
    return process(data)
}

export const getFileArtifactInfo = async (url, fileArtifactName) => {
    const containerClient = new ApiContainerServicePromiseClient(url);
    const makeGetFileArtifactInfo = async () => {
        try {
            const fileArtifactArgs = new FilesArtifactNameAndUuid()
            fileArtifactArgs.setFilename(fileArtifactName)
            const inspectFileArtifactReq = new InspectFilesArtifactContentsRequest()
            inspectFileArtifactReq.setFileNamesAndUuid(fileArtifactArgs)
            const responseFromGrpc = await containerClient.inspectFilesArtifactContents(inspectFileArtifactReq, {})
            return responseFromGrpc.toObject()
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

        data.fileDescriptionsList.map(file => {
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

    const response = await getDataFromApiContainer(makeGetFileArtifactInfo, processFileArtifact)
    return response;
}

export const getEnclaveInformation = async (url) => {
    if (url === "") {
        return {
            services: [],
            artifacts: []
        }
    }

    const containerClient = new ApiContainerServicePromiseClient(url);
    const makeGetServiceRequest = async () => {
        try {
            return await getServicesFromEnclaveManager();
        } catch (error) {
            return {serviceInfo: []}
        }
    }

    const makeFileArtifactRequest = async () => {
        const fileArtifactResponse = await containerClient.listFilesArtifactNamesAndUuids(new google_protobuf_empty_pb.Empty, null)
        return fileArtifactResponse.toObject();
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
        return data.fileNamesAndUuidsList.map(artifact => {
            return {
                name: artifact.filename,
                uuid: artifact.fileuuid,
            }
        })
    }

    const servicesPromise = getDataFromApiContainer(makeGetServiceRequest, processServiceRequest)
    const fileArtifactsPromise = getDataFromApiContainer(makeFileArtifactRequest, processFileArtifactRequest)

    const [services, artifacts] = await Promise.all([servicesPromise, fileArtifactsPromise])
    console.log("sa", services, artifacts)
    return {services, artifacts}
}

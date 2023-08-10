import google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb.js'
import {GetServicesArgs, RunStarlarkPackageArgs, FilesArtifactNameAndUuid, InspectFilesArtifactContentsRequest} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import {ApiContainerServicePromiseClient} from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb'

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

export const getFileArtifactInfo = async(url, fileArtifactName) => {
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
                files:[]
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
            const serviceArgs = new GetServicesArgs();
            const responseFromGrpc = await containerClient.getServices(serviceArgs, null)
            return responseFromGrpc.toObject()
        } catch (error) {
            return {serviceInfoMap:[]}
        }
    }

    const makeFileArtifactRequest = async () => {
        const fileArtifactResponse = await containerClient.listFilesArtifactNamesAndUuids(new google_protobuf_empty_pb.Empty, null)
        return fileArtifactResponse.toObject();      
    }
    
    const processServiceRequest = (data) => {
        return data.serviceInfoMap.map(service => {
            const ports = service[1].maybePublicPortsMap.map((publicPort, index) => {
                const privatePort = service[1].privatePortsMap[index]; 
                return {
                    publicPortNumber:publicPort[1].number,
                    privatePortNumber: privatePort[1].number,
                    applicationProtocol: privatePort[1].maybeApplicationProtocol,
                    portName: privatePort[0],
                    transportProtocol: TransportProtocolEnum[privatePort[1].transportProtocol]
                }
            })
    
            return {
                name: service[0],
                uuid: service[1].serviceUuid,
                privateIpAddr: service[1].privateIpAddr,
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
    return { services, artifacts}
}

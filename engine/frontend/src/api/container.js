import google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb.js'
import {GetServicesArgs, RunStarlarkPackageArgs} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import {ApiContainerServicePromiseClient} from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb'

const TransportProtocolEnum = ["tcp", "sctp", "udp"];

export const runStarlarkPackage = async (url, packageId) => {
    const containerClient = new ApiContainerServicePromiseClient(url);
    const runStarlarkPackageArgs = new RunStarlarkPackageArgs();

    runStarlarkPackageArgs.setDryRun(false);
    runStarlarkPackageArgs.setRemote(true);
    runStarlarkPackageArgs.setPackageId(packageId);
    runStarlarkPackageArgs.setSerializedParams("{}")
    const stream = containerClient.runStarlarkPackage(runStarlarkPackageArgs, null);
    return stream;
}

export const getEnclaveInformation = async (url) => {
    const containerClient = new ApiContainerServicePromiseClient(url);
    const serviceArgs = new GetServicesArgs();
    const responseFromGrpc = await containerClient.getServices(serviceArgs, null)
    const response = responseFromGrpc.toObject();
    const services = response.serviceInfoMap.map(service => {
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

    return services;
}
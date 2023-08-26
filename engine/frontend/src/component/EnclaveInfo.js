import Heading from "./Heading";
import {useEffect, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import {getEnclaveInformation} from "../api/container";

import NoData from "./NoData";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import LoadingOverlay from "./LoadingOverflow";
import {useAppContext} from "../context/AppState";

const renderEnclaves = (enclaves, handleClick) => {
    return enclaves.map(enclave => {
        const backgroundColor = enclave.status === 1 ? "bg-[#24BA27]" : "bg-red-600"
        return (
            <div className={`flex items-center justify-center h-14 text-base ${backgroundColor}`} key={enclave.name}
                 onClick={() => handleClick(enclave.name)}>
                <div className='cursor-default text-lg text-white'> {enclave.name} </div>
            </div>
        )
    })
}

const renderServices = (services, handleClick) => {
    if (services.length === 0) {
        return (
            <NoData
                text={`There are no running services in this enclave`}
                size={`text-xl`}
                color={`text-red-400`}
            />
        )
    }

    return services.map((service) => {
        return (
            <div className="border-4 bg-[#171923] text-lg align-middle text-center h-16 p-3 text-[#24BA27]"
                 onClick={() => handleClick(service, services)}>
                <div> {service.name} </div>
            </div>
        )
    })
}

const renderFileArtifacts = (file_artifacts, handleFileArtifactClick) => {
    if (file_artifacts.length === 0) {
        return (
            <NoData
                text={`There are no file artifacts in this enclave`}
                size={`text-xl`}
                color={`text-red-400`}
            />
        )
    }

    return file_artifacts.map((file_artifact) => {
        return (
            <div className="border-4 bg-[#171923] text-lg align-middle text-center h-16 p-3 text-[#24BA27]"
                 onClick={() => handleFileArtifactClick(file_artifact.name, file_artifacts)}>
                <div>{file_artifact.name}</div>
            </div>
        )
    })
}

const EncalveInfo = ({enclaves}) => {
    const navigate = useNavigate();
    const {appData} = useAppContext()

    const params = useParams();
    const {name} = params;

    const [services, setServices] = useState([])
    const [fileArtifacts, setFileArtifacts] = useState([])
    const [encalveInfoLoading, setEnclaveInfoLoading] = useState(false)

    useEffect(() => {
        // console.log("EnclaveInfo: ", enclaves)
        setEnclaveInfoLoading(true)
        const fetch = async () => {
            const selected = enclaves.filter(enclave => enclave.name === name);
            if (selected.length > 0) {
                const {
                    services,
                    artifacts
                } = await getEnclaveInformation(selected[0].host, selected[0].port, appData.jwtToken, appData.apiHost);
                setServices(services)
                setFileArtifacts(artifacts)
            }
            setEnclaveInfoLoading(false)
        }
        fetch()
    }, [name, enclaves])

    const handleServiceClick = (service, services) => {
        navigate(`services/${service.uuid}`, {state: {services, selected: service}})
    }

    const handleLeftPanelClick = (enclaveName) => {
        navigate(`./../${enclaveName}`, {replace: true})
    }

    const handleFileArtifactClick = async (fileArtifactName, fileArtifacts) => {
        // console.log("Artifacts: ", fileArtifacts)
        const selected = enclaves.filter(enclave => enclave.name === name);
        if (selected.length > 0) {
            navigate(`enclaves/${selected[0].name}/files/${fileArtifactName}`, {state: {fileArtifacts}})
        }
    }

    const EnclaveInfoComponent = ({services, fileArtifacts, handleServiceClick, handleFileArtifactClick}) => (
        <div className='flex flex-col h-[calc(100vh-3rem)] space-y-1 overflow-auto'>
            <div className="flex flex-col h-1/2 min-h-1/2 border-8">
                <Heading content={"Services"} size={"text-xl"}/>
                <div className="overflow-auto space-y-2">
                    {renderServices(services, handleServiceClick)}
                </div>
            </div>
            <div className="flex flex-col h-[46%] border-8">
                <Heading content={"File Artifacts"} size={"text-xl"} padding={"p-1"}/>
                <div className="overflow-auto space-y-2">
                    {renderFileArtifacts(fileArtifacts, handleFileArtifactClick)}
                </div>
            </div>
        </div>
    )

    return (
        <div className="flex h-full">
            <LeftPanel
                home={false}
                heading={"Enclaves"}
                renderList={() => renderEnclaves(enclaves, handleLeftPanelClick)}
            />

            <div className="flex bg-white w-[calc(100vw-39rem)] flex-col space-y-5">
                <div className="h-[3rem] flex items-center justify-center m-2">
                    <Heading content={name}/>
                </div>
                {encalveInfoLoading ?
                    <LoadingOverlay/> :
                    <EnclaveInfoComponent
                        services={services}
                        fileArtifacts={fileArtifacts}
                        handleServiceClick={handleServiceClick}
                        handleFileArtifactClick={handleFileArtifactClick}
                    />
                }
            </div>

            <RightPanel/>
        </div>
    )
}

export default EncalveInfo;

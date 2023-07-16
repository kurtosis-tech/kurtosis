import Heading from "./Heading";
import { useEffect, useState } from "react";
import {useNavigate, useParams, useLocation} from "react-router-dom";
import {getEnclaveInformation} from "../api/container";

import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";


const renderEnclaves = (enclaves, handleClick) => {
    return enclaves.map(enclave => {
        const backgroundColor = enclave.status === 1 ? "bg-green-700": "bg-red-600"
        return (
            <div className={`flex items-center justify-center h-14 text-base ${backgroundColor}`} key={enclave.name} onClick={()=>handleClick(enclave.name)}>
                <div className='cursor-default text-lg text-white'> {enclave.name} </div>
            </div>
        )
    })
}

const renderServices = (services, handleClick) => {
    if (services.length === 0) {
        return (
            <div className="text-3xl text-red-600 text-center justify-center">
                No Data: 
                This occurs because either enclave is stopped or there was error while executing
                the package.
             </div>
        )
    }

    return services.map((service)=> {
        return (
            <div className="border-4 bg-slate-800 text-lg align-middle text-center h-16 p-3 text-green-600" onClick={() => handleClick(service, services)}> 
                <div> {service.name} </div>
            </div>
        )
    })
}

const renderFileArtifacts = (file_artifacts) => {
    if (file_artifacts.length === 0) {
        return (
            <div className="text-3xl text-slate-200 text-center justify-center">
                No Data
             </div>
        )
    }

    return file_artifacts.map((file_artifact)=> {
        return (
            <div className="border-4 bg-slate-800 text-lg align-middle text-center h-16 p-3 text-green-600"> 
                <div> {file_artifact.name} </div>
            </div>
        )
    })
}

const EncalveInfo = ({enclaves}) => {
    const navigate = useNavigate();
    
    const params = useParams();
    const {name} = params;

    const [services, setServices] = useState([])
    const [fileArtifacts, setFileArtifacts] = useState([])
    
    useEffect(() => {
        const fetch = async () => {
            const selected = enclaves.filter(enclave => enclave.name === name);
            
            if (selected.length > 0) {
                const {services, artifacts} = await getEnclaveInformation(selected[0].apiClient);
                setServices(services)
                setFileArtifacts(artifacts)
            }
            
        } 
        fetch()
    }, [name, enclaves])

    const handleServiceClick = (service, services) => {
        navigate(`/enclaves/${name}/services/${service.uuid}`, {state: {services, selected: service}})
    }

    const handleLeftPanelClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`, {replace:true})
    }

    return (
        <div className="flex h-full bg-white">
            <LeftPanel 
                home={false} 
                heading={"Environments"} 
                renderList={ ()=> renderEnclaves(enclaves, handleLeftPanelClick)}
            />

            <div className="flex-1">
                <Heading content={name} />
                <div className='flex flex-col h-full space-y-1'>
                    <div className="flex flex-col h-1/2 min-h-1/2 border-8">
                        <Heading content={"Services"} size={"text-xl"} />
                        <div className="overflow-auto space-y-2">
                            {renderServices(services, handleServiceClick)}
                        </div>
                    </div>  
                    <div className="flex flex-col overflow-auto h-full border-8">
                        <Heading content={"File Artifacts"} size={"text-xl"} padding={"p-1"}/>
                        <div className="overflow-auto space-y-2">
                            {renderFileArtifacts(fileArtifacts)}
                        </div>
                    </div>  
                </div>
            </div>
                    
            <RightPanel/>
        </div>
    )
}

export default EncalveInfo;
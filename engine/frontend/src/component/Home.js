import React from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import FileArtifactInfo from './FileArtifactInfo';
import Enclaves from "./Enclaves";
import CreateEnclave from "./CreateEnclave"
import {useEffect, useState} from "react";
import {getEnclavesFromKurtosis} from "../api/enclave";

import {Route, Routes} from 'react-router-dom';
import {useAppContext} from "../context/AppState";

const Home = () => {
    const [enclaves, setEnclaves] = useState([])
    const [encalveLoading, setEnclaveLoading] = useState(false)
    const {appData, setAppData} = useAppContext()

    useEffect(() => {
        setEnclaveLoading(true)
        const fetch = async () => {
            const response = await getEnclavesFromKurtosis(appData.jwtToken);
            setEnclaves(response)
            setEnclaveLoading(false)
        } 
        fetch()
    }, [appData.jwtToken])

    const addEnclave = (enclave) => {
        setEnclaves(enclaves => [...enclaves, enclave])
    }

    return (
        <div className="h-screen flex flex-col bg-[#171923]">
            <TitleBar/>
            <div className="flex h-[calc(100vh-4rem)]">
                <Routes>
                    <Route exact path="/" element={<Main totalEnclaves={enclaves.length}/>}/>
                    <Route exact path="/enclave/*" element={<CreateEnclave addEnclave={addEnclave}/>}/>
                    <Route exact path="/enclaves" element={<Enclaves enclaves={enclaves} isLoading={encalveLoading}/>}/>
                    <Route path="/enclaves/:name" element={<EnclaveInfo enclaves={enclaves}/>}/>
                    <Route path="/enclaves/:name/services/:uuid" element={<ServiceInfo/>}/>
                    <Route path="/enclaves/:name/files/:fileArtifactName"
                           element={<FileArtifactInfo enclaves={enclaves}/>}/>
                </Routes>
            </div>
        </div>
    );
}

export default Home;

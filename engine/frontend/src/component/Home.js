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

import {Route, Routes, useSearchParams} from 'react-router-dom';
import {useAppContext} from "../context/AppState";
import LoadingOverlay from "./LoadingOverflow";

const loading = (
    <div className="flex-grow bg-#181926-100 flex-row flex mt-28 w-screen">
        <div className='w-1/3'></div>
        <div className="flex flex-col min-w-fit w-1/3">
            <div className='flex justify-center items-center w-full'>
                <div className='text-center w-full'>
                    <h1 className="text-3xl font-bold text-[#24BA27]">Loading credentials...</h1>
                    <br/>
                    <LoadingOverlay/>
                </div>
            </div>
        </div>
        <div className='w-1/3'></div>
    </div>
)
const queryParamToBool = (value) => {
    return ((value + '').toLowerCase() === 'true')
}

const Home = () => {
    const [enclaves, setEnclaves] = useState([])
    const [enclaveLoading, setEnclaveLoading] = useState(false)
    const {appData, setAppData} = useAppContext()
    const [searchParams, setSearchParams] = useSearchParams();

    const receiveMessage = (event) => {
        const message = event.data.message;
        switch (message) {
            case 'jwtToken':
                const value = event.data.value
                if (value !== null && value !== undefined) {
                    console.log("Got the message!!", value)
                    setAppData({
                        ...appData,
                        jwtToken: value,
                    })
                }
                break;
        }
    }
    window.addEventListener("message", receiveMessage)

    const requireAuth = queryParamToBool(searchParams.get("require_authentication"))

    useEffect(() => {
        if (requireAuth && !appData.jwtToken) {
            console.log("waiting for auth")
        } else {

            setEnclaveLoading(true)
            const fetch = async () => {
                const response = await getEnclavesFromKurtosis(appData.jwtToken);
                setEnclaves(response)
                setEnclaveLoading(false)
            }
            fetch()
        }
    }, [appData.jwtToken])

    const addEnclave = (enclave) => {
        setEnclaves(enclaves => [...enclaves, enclave])
    }

    const checkAuth = (element) => {
        if (!requireAuth) {
            return element;
        }
        if (requireAuth && appData.jwtToken) {
            return element;
        }
        if (requireAuth && !appData.jwtToken) {
            return loading;
        }
    }

    return (
        <div className="h-screen flex flex-col bg-[#171923]">
            <TitleBar/>
            <div className="flex h-[calc(100vh-4rem)]">
                <Routes>
                    <Route exact path="/" element={checkAuth(<Main totalEnclaves={enclaves.length}/>)}/>
                    <Route exact path="/enclave/*" element={checkAuth(<CreateEnclave addEnclave={addEnclave}/>)}/>
                    <Route exact path="/enclaves"
                           element={checkAuth(<Enclaves enclaves={enclaves} isLoading={enclaveLoading}/>)}/>
                    <Route path="/enclaves/:name" element={checkAuth(<EnclaveInfo enclaves={enclaves}/>)}/>
                    <Route path="/enclaves/:name/services/:uuid" element={checkAuth(<ServiceInfo/>)}/>
                    <Route path="/enclaves/:name/files/:fileArtifactName"
                           element={checkAuth(<FileArtifactInfo enclaves={enclaves}/>)}/>
                </Routes>
            </div>
        </div>
    );
}

export default Home;

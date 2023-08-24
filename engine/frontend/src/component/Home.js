import React, {useEffect, useState} from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import FileArtifactInfo from './FileArtifactInfo';
import Enclaves from "./Enclaves";
import CreateEnclave from "./CreateEnclave"
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

const DefaultApiHost = "localhost"
const DefaultApiPort = 8081

const createApiUrl = (apiHost, requireHttps) => {
    if (requireHttps) {
        return `https://cloud.kurtosis.com/gateway/ips/${apiHost}/ports/${DefaultApiPort}`;
    }
    return `http://${DefaultApiHost}:${DefaultApiPort}`;
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
        // At this time requireAuth=true means we are running remote which means connection is going through a TLS protected LB
        const requireHttps = requireAuth;
        const apiHost = createApiUrl(searchParams.get("api_host"), requireHttps)
        if (apiHost && apiHost.length > 0) {
            console.log(`Setting API host = ${apiHost}`)
            setAppData({
                ...appData,
                apiHost: apiHost,
            })
        } else {
            console.error("Could not determine the api host.")
        }
    }, [appData.apiHost])

    useEffect(() => {
        if (requireAuth && !appData.jwtToken) {
            console.log("Waiting for auth token")
        } else {
            if (appData.jwtToken) {
                console.log("Got auth token")
            }
            if (appData.apiHost) {
                setEnclaveLoading(true)
                const fetch = async () => {
                    const response = await getEnclavesFromKurtosis(appData.jwtToken, appData.apiHost);
                    setEnclaves(response)
                    setEnclaveLoading(false)
                }
                fetch()
            }
        }
    }, [appData.jwtToken, appData.apiHost])

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

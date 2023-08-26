import React, {useEffect, useMemo, useState} from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import FileArtifactInfo from './FileArtifactInfo';
import Enclaves from "./Enclaves";
import {getEnclavesFromKurtosis} from "../api/enclave";

import {Route, Routes, useSearchParams} from 'react-router-dom';
import {useAppContext} from "../context/AppState";
import LoadingOverlay from "./LoadingOverflow";
import CreateEnclave from "./CreateEnclave";


const queryParamToBool = (value) => {
    return ((value + '').toLowerCase() === 'true')
}

const DefaultApiHost = "localhost"
const DefaultApiPort = 8081

const createApiUrl = (apiHost, requireProxy) => {
    if (requireProxy) {
        return `https://cloud.kurtosis.com/gateway/ips/${apiHost}/ports/${DefaultApiPort}`;
    }
    return `http://${DefaultApiHost}:${DefaultApiPort}`;
}

const makeUrl = () => {
    let path = window.location.pathname
    if (path.charAt(0) === "/") path = path.substr(1);
    path = path.endsWith('/') ? path.slice(0, -1) : path;
    if (!path) return ""
    return path;
}
const Home = () => {
    const [enclaves, setEnclaves] = useState([])
    const [enclaveLoading, setEnclaveLoading] = useState(false)
    const {appData, setAppData} = useAppContext()
    const [searchParams, setSearchParams] = useSearchParams();

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

    const urlPath = useMemo(() => {
        return makeUrl()
    }, [])


    const receiveMessage = (event) => {
        const message = event.data.message;
        switch (message) {
            case 'jwtToken':
                const value = event.data.value
                console.log("got event: jwtToken")
                if (value !== null && value !== undefined) {
                    // console.log("got token:", value)
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
        const requestedApiHost = searchParams.get("api_host")
        const requireProxy = requireAuth;
        const apiHost = createApiUrl(requestedApiHost, requireProxy)
        if (apiHost && apiHost.length > 0) {
            setAppData({
                ...appData,
                apiHost: apiHost,
            })
        }
    }, [])

    const validJwtToken = () => appData.jwtToken && appData.jwtToken.length > 0
    const validApiHost = () => appData.apiHost && appData.apiHost.length > 0

    const fetch = async () => {
        console.log("submitting request for enclaves")
        const response = await getEnclavesFromKurtosis(appData.jwtToken, appData.apiHost);
        console.log("Got response for enclaves", response)
        setEnclaves(response)
        setEnclaveLoading(false)
        console.log("finished fetch")
    }

    useEffect(() => {
        if (requireAuth && !validJwtToken()) {
            console.log("Requires Auth and jwt token: waiting for jwt token")
        }
        if (requireAuth && validJwtToken() && validApiHost()) {
            console.log("Requires Auth, jwt token and api host: Got them all")
            console.log("starting load for authenticated access")
            setEnclaveLoading(true)
            fetch()
        }

    }, [appData.jwtToken, appData.apiHost])

    useEffect(() => {
        if (!requireAuth && validApiHost()) {
            console.log("Does not require auth. Got Api host:", appData.apiHost)
            console.log("starting load for non authenticated access")
            fetch()
        }
    }, [appData.apiHost])

    const addEnclave = (enclave) => {
        setEnclaves(enclaves => [...enclaves, enclave])
    }

    const checkAuth = (element) => {
        if (!requireAuth && appData.apiHost) {
            return element;
        }
        if (requireAuth && appData.jwtToken && appData.apiHost) {
            return element;
        }
        if (requireAuth && (!appData.jwtToken || appData.apiHost)) {
            return loading;
        }
        return element;
    }

    const constructPath = (path) => {
        return `${urlPath}${path}`
    }
    return (
        <div className="h-screen flex flex-col bg-[#171923]">
            <TitleBar/>
            <div className="flex h-[calc(100vh-4rem)]">
                <Routes>
                    <Route exact
                           path={constructPath("/")}
                           element={checkAuth(<Main totalEnclaves={enclaves.length}/>)}
                    />
                    <Route exact
                           path={constructPath("/enclaves")}
                           element={checkAuth(<Enclaves enclaves={enclaves}
                                                        isLoading={enclaveLoading}
                               />
                           )}
                    />
                    <Route exact
                           path={constructPath("/enclave/*")}
                           element={checkAuth(<CreateEnclave addEnclave={addEnclave}/>)}
                    />
                    <Route path={constructPath("/enclaves/:name")}
                           element={checkAuth(<EnclaveInfo enclaves={enclaves}/>)}
                    />
                    <Route path={constructPath("/enclaves/:name/services/:uuid")}
                           element={checkAuth(<ServiceInfo/>)}
                    />
                    <Route path={constructPath("/enclaves/:name/files/:fileArtifactName")}
                           element={checkAuth(<FileArtifactInfo enclaves={enclaves}/>)}
                    />
                </Routes>
            </div>
        </div>
    );
}

export default Home;

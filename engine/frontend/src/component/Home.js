import React, {useEffect, useMemo, useState} from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import FileArtifactInfo from './FileArtifactInfo';
import PackageCatalogRouter from './PackageCatalogRouter';
import Enclaves from "./Enclaves";
import {getEnclavesFromKurtosis, removeEnclave} from "../api/enclave";
import {createBrowserRouter, Route, RouterProvider} from 'react-router-dom';
import {useAppContext} from "../context/AppState";
import LoadingOverlay from "./LoadingOverflow";
import CreateEnclave from "./CreateEnclave";
import {createRoutesFromElements, Outlet} from "react-router";

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
    if (!path || path.length === 0) return "/"
    if (path.charAt(0) !== "/") path = path + "/"
    return path;
}

const Layout = () => {
    return (
        <>
            <TitleBar/>
            <div className="flex h-[calc(100vh-4rem)]">
                <Outlet/>
            </div>
        </>
    )
}

const Home = () => {
    const [enclaves, setEnclaves] = useState([])
    const [enclaveLoading, setEnclaveLoading] = useState(false)
    const {appData, setAppData} = useAppContext()

    const handleDeleteClick = async (enclaveName) => {
        const makeRequest = async () => {
            try {
                const filteredEnclaves = enclaves.filter(enclave => enclave.name !== enclaveName)
                await removeEnclave(appData.jwtToken, appData.apiHost, enclaveName)
                setEnclaves(filteredEnclaves)
            } catch (ex) {
                console.error(ex)
                alert(`Sorry, unexpected error occurred while removing enclave with name: ${enclaveName}`)
            }
        }
        await makeRequest()
    }

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

    // Using plain JS for now since useLocation requires to be inside Router.
    const searchParams = new URLSearchParams(window.location.search);
    const requireAuth = queryParamToBool(searchParams.get("require_authentication"))
    const requestedApiHost = searchParams.get("api_host")
    const preloadedPackage = searchParams.get("package")

    useEffect(() => {
        // At this time requireAuth=true means we are running remote which means connection is going through a TLS protected LB
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
        const response = await getEnclavesFromKurtosis(appData.jwtToken, appData.apiHost);
        const filteredResponse = response.filter((e) => {
            return e.name
        });
        setEnclaves(filteredResponse)
        setEnclaveLoading(false)
    }

    useEffect(() => {
        if (requireAuth && validJwtToken() && validApiHost()) {
            setEnclaveLoading(true)
            fetch()
        }

    }, [appData.jwtToken, appData.apiHost])

    useEffect(() => {
        if (!requireAuth && validApiHost()) {
            fetch()
        }
    }, [appData.apiHost])

    const addEnclave = (enclave) => {
        const {created, ...updated} = enclave
        setEnclaves(enclaves => [...enclaves, updated])
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

    const routes = (
        <>
            <Route path="/" element={checkAuth(<Layout/>)}>
                <Route
                    path="/enclaves"
                    element={<Enclaves
                        enclaves={enclaves}
                        isLoading={enclaveLoading}
                        handleDeleteClick={handleDeleteClick}
                    />
                    }
                />
                <Route
                    path="/enclave/*"
                    element={<CreateEnclave addEnclave={addEnclave}/>}
                />
                <Route path="/enclaves/:name"
                       element={<EnclaveInfo enclaves={enclaves}/>}
                />
                <Route path="/enclaves/:name/services/:uuid/*"
                       element={<ServiceInfo/>}
                />
                <Route path="/enclaves/:name/files/:fileArtifactName"
                       element={<FileArtifactInfo enclaves={enclaves}/>}
                />
                <Route
                    path="/catalog/*"
                    element={<PackageCatalogRouter addEnclave={addEnclave}/>}
                />
                <Route exact
                       path="/"
                       element={<Main totalEnclaves={enclaves.length} preloadedPackage={preloadedPackage}/>}
                />
            </Route>
        </>
    )

    const router = createBrowserRouter(
        createRoutesFromElements(routes),
        {
            basename: urlPath,
        }
    );

    return (
        <div className="h-screen flex flex-col bg-[#171923]">
            <RouterProvider router={router}/>
        </div>
    );
}

export default Home;

import { Route, Routes, useNavigate } from 'react-router-dom';
import PackageCatalogForm from './PackageCatalogForm';
import PackageCatalog from './PackageCatalog';
import PackageCatalogProgress from "./PackageCatalogProgress";
import { useEffect } from 'react';
import {getKurtosisPackages} from "../api/packageCatalog";
import {createEnclave} from "../api/enclave";
import {useState} from "react";
import {useAppContext} from "../context/AppState";


const PackageCatalogRouter = ({addEnclave}) => {
    const navigate = useNavigate()
    const {appData} = useAppContext()
    const [kurtosisPackages, setKurtosisPackages] = useState([])

    useEffect(()=> {
        const fetchPackages = async () => {
            const resp = await getKurtosisPackages()
            setKurtosisPackages(resp)
        }
        fetchPackages();
    },[])

    const createNewEnclave = async (runArgs, enclaveName, productionMode) => {
        const request = async () => {
            try {
                const enclave = await createEnclave(appData.jwtToken, appData.apiHost, enclaveName, productionMode);
                addEnclave(enclave)
                navigate("/catalog/progress", {state: {
                    enclave,
                    runArgs,
                }})
            } catch(ex) {
                console.log(ex)
                alert(`Error occurred while creating enclave for package: ${runArgs.kurtosisPackage.name}. An error message should be printed in console, please share it with us to help debug this problem`)
            } 
        }
        await request()
    }

    return (
        <div className='h-full w-full flex'>
            <Routes>
                <Route path="/progress" element={
                    <PackageCatalogProgress appData={appData}/>
                } />
                <Route path="/form" element={
                    <PackageCatalogForm createEnclave={createNewEnclave}/>
                } />
                <Route path="/" element={
                    <PackageCatalog kurtosisPackages={kurtosisPackages}
                />}/>
            </Routes>
        </div>
    )
}

export default PackageCatalogRouter;
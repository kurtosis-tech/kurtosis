import { Route, Routes, useNavigate } from 'react-router-dom';
import PackageCatalogForm from './PackageCatalogForm';
import PackageCatalog from './PackageCatalog';
import PackageCatalogProgress from "./PackageCatalogProgress";
import { useEffect } from 'react';
import {getKurtosisPackages} from "../api/enclave";
import {createEnclave} from "../api/enclave";
import {useState} from "react";

const PackageCatalogRouter = () => {
    const navigate = useNavigate()
    const [kurtosisPackages, setKurtosisPackages] = useState([])
    
    useEffect(()=> {
        const fetchPackages = async () => {
            const resp = await getKurtosisPackages()
            setKurtosisPackages(resp)
        }
        fetchPackages();
    },[])

    const createNewEnclave = (runArgs) => {
        const request = async () => {
            const enclave = await createEnclave();
            navigate("/catalog/progress", {state: {
                enclave,
                runArgs,
            }})
        }
        request()
    }

    return (
        <div className='h-full w-full flex'>
            <Routes>
                <Route path="/progress" element={
                    <PackageCatalogProgress/>
                } />
                <Route path="/form" element={
                    <PackageCatalogForm handleCreateNewEnclave={createNewEnclave}/>
                } />
                <Route path="/" element={
                    <PackageCatalog kurtosisPackages={kurtosisPackages}
                />}/>
            </Routes>
        </div>
    )
}

export default PackageCatalogRouter;
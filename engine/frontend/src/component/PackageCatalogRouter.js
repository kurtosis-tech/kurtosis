import { Route, Routes } from 'react-router-dom';
import PackageCatalogForm from './PackageCatalogForm';
import PackageCatalog from './PackageCatalog';
import { useEffect } from 'react';
import {getKurtosisPackages} from "../api/enclave";
import {useState} from "react";

const PackageCatalogRouter = () => {
    const [kurtosisPackages, setKurtosisPackages] = useState([])
    useEffect(()=> {
        const fetchPackages = async () => {
            const resp = await getKurtosisPackages()
            setKurtosisPackages(resp)
        }
        fetchPackages();
    },[])

    return (
        <div className='h-full w-full flex'>
            <Routes>
                <Route path="/progress" element={
                    <PackageCatalogForm/>
                } />
                <Route path="/" element={
                    <PackageCatalog kurtosisPackages={kurtosisPackages}
                />}/>
            </Routes>
        </div>
    )
}

export default PackageCatalogRouter;
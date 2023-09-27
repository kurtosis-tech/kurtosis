import React, {useState} from 'react';

import {CreateEnclaveModal} from "./CreateEnclaveModal";
import {CreateEnclaveLog} from "./log/CreateEnclaveLog";
import {Route, Routes, useNavigate} from 'react-router-dom';
import {useAppContext} from "../context/AppState";

const CreateEnclave = ({addEnclave}) => {
    const navigate = useNavigate()
    const [enclaveName, setEnclaveName] = useState("")
    const [name, setName] = useState('');
    const [args, setArgs] = useState('{}')
    const [enclave, setEnlave] = useState(null);
    const [productionMode, setProductionMode] = useState(false)

    const {appData} = useAppContext()

    const handleModalSubmit = (enclave) => {
        setEnlave(enclave)
        navigate("/enclave/progress")
    }

    return (
        <div className='h-full w-full flex'>
            <Routes>
                <Route path="/create"
                       element={<CreateEnclaveModal enclaveName={enclaveName} addEnclave={addEnclave} name={name} setName={setName} setEnclaveName={setEnclaveName} args={args}
                                                    setArgs={setArgs} handleSubmit={handleModalSubmit}
                                                    token={appData.jwtToken}
                                                    apiHost={appData.apiHost}
                                                    productionMode={productionMode}
                                                    setProductionMode={setProductionMode}
                                                    />}/>
                <Route path="/progress" element={
                    <CreateEnclaveLog 
                        args={args} 
                        packageId={name} 
                        enclave={enclave} 
                        appData={appData}
                    />}/>
            </Routes>
        </div>
    )
}

export default CreateEnclave;

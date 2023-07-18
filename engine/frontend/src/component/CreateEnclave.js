import React, { useState} from 'react';

import {CreateEnclaveModal} from "./CreateEnclaveModal";
import {CreateEnclaveView} from "./CreateEnclaveView";
import { Route, Routes, useNavigate } from 'react-router-dom';

const CreateEnclave = () => {
    const navigate = useNavigate()
    const [name, setName] = useState('');
    const [args, setArgs] = useState('{}')
    const [enclaveInfo, setEnlaveInfo] = useState(null);

    const handleModalSubmit = (enclaveInfo) => {
        setEnlaveInfo(enclaveInfo)
        //navigate("/enclave/progress")
    }

    return (
        <div className='h-full w-full flex'>
            {/* <Routes>
                <Route path="/create" element={></Route>
                <Route path="/progress" element={}></Route>
            </Routes> */}
            {enclaveInfo !== null ? <CreateEnclaveView args={args} packageId={name} enclaveInfo={enclaveInfo}/>
            :<CreateEnclaveModal name={name} setName={setName} args={args} setArgs={setArgs} handleSubmit={handleModalSubmit}/>}
        </div>
    )
}

export default CreateEnclave;
import React, { useState } from 'react';

import {CreateEnclaveModal} from "./CreateEnclaveModal";
import {CreateEnclaveView} from "./CreateEnclaveView";

const CreateEnclave = () => {
    const [name, setName] = useState('');
    const [enclaveInfo, setEnlaveInfo] = useState(null);

    const handleModalSubmit = (enclaveInfo) => {
        setEnlaveInfo(enclaveInfo)
    }

    return (
        <div className='h-screen w-screen'>
            {enclaveInfo !== null ? <CreateEnclaveView packageId={name} enclaveInfo={enclaveInfo}/> : <CreateEnclaveModal name={name} setName={setName} handleSubmit={handleModalSubmit}/>}
        </div>
    )
}

export default CreateEnclave;
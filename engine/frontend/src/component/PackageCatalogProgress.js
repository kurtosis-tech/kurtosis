import {CreateEnclaveLog} from "./log/CreateEnclaveLog";
import {useLocation} from "react-router-dom";

const PackageCatalogProgress = ({appData}) => {
    const location = useLocation()
    const {state} = location;
    const enclave = state.enclave;
    const packageId = state.runArgs.packageId;
    const args = state.runArgs.args;

    console.log("state", state)

    return (
        <div className='h-full w-full flex'>
            <CreateEnclaveLog 
                args={args} 
                packageId={packageId} 
                enclave={enclave} 
                appData={appData}
            />
        </div>
    )
}

export default PackageCatalogProgress;

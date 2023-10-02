import {useEffect, useState} from "react";
import {Route, Routes, useLocation, useNavigate, useParams} from "react-router-dom";
import {Log} from "./log/Log";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import {getServiceLogs} from "../api/enclave";
import {useAppContext} from "../context/AppState";
import {Box, Button, Center, Flex, Spacer, Text} from "@chakra-ui/react";
import ServiceView from "./ServiceView";

const DEFAULT_SHOULD_FOLLOW_LOGS = true
const DEFAULT_NUM_LINES = 1500

const renderServices = (services, handleClick) => {
    return services.map(service => {
        return (
            <div className={`flex items-center justify-center h-14 text-base bg-[#24BA27]`} key={service.name}
                 onClick={() => handleClick(service)}>
                <div className='cursor-default text-lg text-white'> {service.name} </div>
            </div>
        )
    })
}

const ServiceInfo = () => {
    const navigate = useNavigate();
    const [viewLog, setViewLog] = useState(false)
    const [logs, setLogs] = useState([])
    const {state} = useLocation();
    const {services, selected} = state;
    const {appData} = useAppContext()

    const params = useParams()
    const {name: enclaveName, uuid: serviceUuid} = params;

    const startServiceLogCollector = (ctrl) => {
        let stream
        const fetchServiceLogs = async () => {
            stream = await getServiceLogs(ctrl, enclaveName, serviceUuid, appData.apiHost, DEFAULT_SHOULD_FOLLOW_LOGS, DEFAULT_NUM_LINES);
            try {
                for await (const res of stream) {
                    const log = res["serviceLogsByServiceUuid"][serviceUuid]["line"][0]
                    if (log !== "") {
                        setLogs(logs => [...logs, log])
                    } 
                }
            } catch (ex) {
                console.log("Abort Initial Log Stream! with error ", ex)
            }
        }

        fetchServiceLogs()
    }

    useEffect(() => {

        const ctrl = new AbortController();
        startServiceLogCollector(ctrl)

        return () => {
            setLogs([])
            ctrl.abort()
        };
    }, [serviceUuid])

    const handleServiceClick = (service) => {
        const fullPath = `/enclaves/${enclaveName}/services/${service.serviceUuid}`
        navigate(fullPath, {state: {services, selected: service}})
    }

    const renderLogView = () => {
        return (
            (logs.length > 0) ? <Log logs={logs} fileName={`${enclaveName}-${selected.name}`} />: <Center color="white"> No Logs Available</Center>
        )
    }

    const isViewLogPage = () => {
        const log = params["*"]
        return log === "logs"
    }

    const switchServiceInfoView = () => {
        if (isViewLogPage()) {
            navigate(`/enclaves/${enclaveName}/services/${selected.serviceUuid}`,  {state: {services, selected: selected}})
        } else {
            navigate(`/enclaves/${enclaveName}/services/${selected.serviceUuid}/logs`,  {state: {services, selected: selected}})
        }
    }

    return (
        <div className="flex h-full">
            <LeftPanel
                home={false}
                isServiceInfo={true}
                heading={"Services"}
                renderList={() => renderServices(services, handleServiceClick)}
            />
            <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                <div className={`flex-col flex space-y-1 bg-[#171923]`}>
                    <Flex bg={"#171923"} height={"80px"}>
                        <Box p='2' m="4">
                            <Text color={"white"} fontSize='xl' as='b'> {!isViewLogPage() ? "Detailed Info " : "Logs "}  for {selected.name} </Text>
                        </Box>
                        <Spacer/>
                        <Button m="4" onClick={switchServiceInfoView}> {!isViewLogPage() ? "Logs" : `Service`} </Button>
                    </Flex>
                    <Routes>
                        <Route path="/logs" element={renderLogView()} />
                        <Route path="/" element={<ServiceView service={selected}/>} />
                    </Routes>
                </div>
            </div>
            <RightPanel home={false} isServiceInfo={true} enclaveName={enclaveName}/>
        </div>
    )
}

export default ServiceInfo;

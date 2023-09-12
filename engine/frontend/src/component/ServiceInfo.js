import {useEffect, useState} from "react";
import {Route, Routes, useLocation, useNavigate, useParams} from "react-router-dom";
import {Log} from "./log/Log";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import {getServiceLogs} from "../api/enclave";
import {useAppContext} from "../context/AppState";
import {Badge, Box, GridItem, Table, TableContainer, Tbody, Td, Tr, Text, Flex, Spacer, Button, Center} from "@chakra-ui/react";
import {CodeEditor} from "./CodeEditor";

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
        const fetchLogs = async () => {
            stream = await getServiceLogs(ctrl, enclaveName, serviceUuid, appData.apiHost, true, 6000);
            try {
                for await (const res of stream) {
                    const log = res["serviceLogsByServiceUuid"][serviceUuid]["line"][0]
                    if (log !== "") {
                        setLogs(logs => [...logs, log])
                    } 
                }
            } catch (ex) {
                console.log("Abort Initial Log Stream!")
            }
        }

        fetchLogs()
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



    const func = () => {
    }

    const codeBox = (id, parameterName, data) => {
        const serializedData = JSON.stringify(data, null, 2)
        return (
            <Box>
                {
                    CodeEditor(
                        func,
                        true,
                        `${parameterName}.json`,
                        ["json"],
                        250,
                        serializedData,
                        true,
                        false,
                        id,
                        true,
                        true,
                        false,
                        "xs",
                        "1px"
                    )
                }
            </Box>
        );
    }

    const tableRow = (heading, data) => {
        let displayData = ""
        try {
            displayData = data()
        } catch (e) {
            console.error("Error while processing row", e)
            displayData = "Error while retrieving information"
        }
        return (
            <Tr key={heading} className="text-white">
                <Td><p><b>{heading}</b></p></Td>
                <Td>{displayData}</Td>
            </Tr>
        );
    };

    const selectedSerialized = selected; // JSON.parse(JSON.stringify(selected))
    const statusBadge = (status) => {
        let color = ""
        let text = ""
        if(status === "RUNNING" || status === 1){
            color="green"
            text="RUNNING"
        } else if (status === "STOPPED" || status === 0) {
            color="red"
            text="STOPPED"
        } else if(status ==="UNKNOWN" || status === 2){
            color = "yellow"
            text="UNKNOWN"
        } else {
            return (
                <Badge>N/A</Badge>
            )
        }
        return (
            <Badge colorScheme={color}>{text}</Badge>
        )
    }

    const renderLogView = () => {
        return (
            (logs.length > 0) ? <Log logs={logs} fileName={`${enclaveName}-${selectedSerialized.name}`} />: <Center color="white"> No Logs Available</Center>
        )
    }

    const isViewLogPage = () => {
        const log = params["*"]
        return log === "logs"
    }

    const switchServiceInfoView = () => {
        if (isViewLogPage()) {
            navigate(`/enclaves/${enclaveName}/services/${selectedSerialized.serviceUuid}`,  {state: {services, selected: selected}})
        } else {
            navigate(`/enclaves/${enclaveName}/services/${selectedSerialized.serviceUuid}/logs`,  {state: {services, selected: selected}})
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
                <div className={`flex-col flex h-full space-y-1 bg-[#171923]`}>
                    <Flex className={"bg-[#171923]"} style={{height: `80px`}}>    
                        <Box p='2' m="4"> 
                            <Text color={"white"} fontSize='xl' as='b'> {!isViewLogPage() ? "Detailed Info " : "Logs "}  for {selectedSerialized.name} </Text>
                        </Box>
                        <Spacer/>
                        <Button m="4" onClick={switchServiceInfoView}> {!isViewLogPage() ? "View Logs" : `View ${selectedSerialized.name}`} </Button>
                    </Flex>
                    <Routes>
                        <Route path="/logs" element={
                            renderLogView()
                        } />
                        <Route path="/" element={
                            <TableContainer>
                                <Table variant='simple' size='sm'>
                                    <Tbody>
                                        {tableRow("Name", () => selectedSerialized.name)}
                                        {tableRow("UUID", () => <pre>{selectedSerialized.serviceUuid}</pre>)}
                                        {tableRow("Status", () => statusBadge(selectedSerialized.serviceStatus))}
                                        {tableRow("Image", () => selectedSerialized.container.imageName)}
                                        {tableRow("Ports", () => codeBox(0, "ports", selectedSerialized.ports))}
                                        {tableRow("ENTRYPOINT", () => codeBox(1, "entrypoint", selectedSerialized.container.entrypointArgs))}
                                        {tableRow("CMD", () => codeBox(2, "cmd", selectedSerialized.container.cmdArgs))}
                                        {tableRow("ENV", () => codeBox(3, "env", selectedSerialized.container.envVars))}
                                    </Tbody>
                                </Table>
                            </TableContainer>
                        } />
                    </Routes>
                </div>
            </div>
            <RightPanel home={false} isServiceInfo={true} enclaveName={enclaveName}/>
        </div>
    )
}

export default ServiceInfo;

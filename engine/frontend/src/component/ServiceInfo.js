import Heading from "./Heading";
import {useEffect, useState} from "react";
import {useLocation, useNavigate, useParams} from "react-router-dom";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import {getServiceLogs} from "../api/enclave";
import {useAppContext} from "../context/AppState";
import {Badge, Box, GridItem, Table, TableContainer, Tbody, Td, Tr} from "@chakra-ui/react";
import {CodeEditor} from "./CodeEditor";
import {LogView} from "./LogView";

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
    const [logs, setLogs] = useState([])
    const {state} = useLocation();
    const {services, selected} = state;
    const {appData} = useAppContext()

    const params = useParams()
    const {name: enclaveName, uuid: serviceUuid} = params;

    useEffect(() => {
        let stream;
        const ctrl = new AbortController();
        const fetchLogs = async () => {
            stream = await getServiceLogs(ctrl, enclaveName, serviceUuid, appData.apiHost);
            try {
                for await (const res of stream) {
                    const log = res["serviceLogsByServiceUuid"][serviceUuid]["line"][0]
                    setLogs(logs => [...logs, log])
                }
            } catch (ex) {
                console.log("aborted stream!")
            }
        }
        fetchLogs()
        return () => {
            ctrl.abort()
            setLogs([]);
        };
    }, [serviceUuid])

    const handleServiceClick = (service) => {
        const fullPath = `/enclaves/${enclaveName}/services/${service.uuid}`
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
            <Tr key={heading}>
                <Td><p><b>{heading}</b></p></Td>
                <Td>{displayData}</Td>
            </Tr>
        );
    };

    const selectedSerialized = selected; // JSON.parse(JSON.stringify(selected))
    const statusBadge = (status) => {
        console.log(`status=${status}`)
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

    return (
        <div className="flex h-full">
            <LeftPanel
                home={false}
                isServiceInfo={true}
                heading={"Services"}
                renderList={() => renderServices(services, handleServiceClick)}
            />
            <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                <div className='flex flex-col h-full space-y-1 bg-[#171923]'>
                    <Heading content={`${enclaveName} - ${selected.name}`}/>
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

                    {/*<LogView heading={`Service Logs`} logs={logs}/>*/}
                </div>
            </div>
            <RightPanel home={false} isServiceInfo={true} enclaveName={enclaveName}/>
        </div>
    )
}

export default ServiceInfo;

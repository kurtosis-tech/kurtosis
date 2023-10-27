import RightPanel from "../RightPanel";
import LeftPanel from "../LeftPanel";
import {Log} from "./Log";
import {useNavigate} from "react-router-dom";
import {runStarlark} from "../../api/enclave";
import {getEnclaveInformation} from "../../api/container";
import LoadingOverlay from "../LoadingOverflow";
import {Box, Center, CircularProgress, Flex, Spacer, Text, Tooltip} from "@chakra-ui/react";
import React, {useEffect, useState} from "react";
import {CheckIcon, WarningIcon} from "@chakra-ui/icons";

export const ERROR = "error"
export const RUN_FINISHED_EVENT = "runFinishedEvent"

const SERVICE_IS_ADDED = "added with service";
const INSTRUCTION = "instruction"
const PROGRESS_INFO = "progressInfo"
const INSTRUCTION_RESULT = "instructionResult"
const EXECUTION_IN_PROGRESS = "Execution in progress"
const STARTING_EXECUTION = "Starting execution"
const SCRIPT_RUN_STATUS_PROCESSING_INDETERMINATE = "processing_indeterminate"
const SCRIPT_RUN_STATUS_PROCESSING = "processing"
const SCRIPT_RUN_STATUS_ERROR = "error"
const SCRIPT_RUN_STATUS_SUCCESS = "success"

const logsStatus = (event, progress) => {
    if (event === SCRIPT_RUN_STATUS_SUCCESS) {
        return (
            <>
                <Tooltip label='Success! Script finished successfully'>
                    <CheckIcon boxSize={6} color='green'/>
                </Tooltip>
            </>
        )
    } else if (event === SCRIPT_RUN_STATUS_PROCESSING) {
        return (
            <>
                <Tooltip label='Script is running...'>
                    <CircularProgress size="30px" value={progress} color="green">
                    </CircularProgress>
                </Tooltip>
            </>
        )
    } else if (event === SCRIPT_RUN_STATUS_PROCESSING_INDETERMINATE) {
        return (
            <>
                <Tooltip label='Script is running...'>
                    <CircularProgress size="30px" isIndeterminate color="green"></CircularProgress>
                </Tooltip>
            </>
        )
    } else if (event === SCRIPT_RUN_STATUS_ERROR) {
        return (
            <>
                <Tooltip label='Error! Script finished with an error'>
                    <WarningIcon boxSize={5} color='red'/>
                </Tooltip>
            </>
        )
    } else {
        return (
            <>
            </>
        )
    }
}


export const CreateEnclaveLog = ({packageId, enclave, args, appData}) => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false)
    const [logs, setLogs] = useState([])
    const [logsCurrentExecutionStatus, setLogsCurrentExecutionStatus] = useState(<></>)
    const [logsExecutionStatusText, setLogsExecutionStatusText] = useState("")
    const [services, setServices] = useState([])
    const [logsComponent, setLogsComponent] = useState(<></>)

    const getServices = async (enclave) => {
        const {services: newServices} = await getEnclaveInformation(enclave.host, enclave.port, appData.jwtToken, appData.apiHost);
        if (newServices.length > services.length) {
            setServices(newServices)
        }
    }

    const readStreamData = (result) => {
        if (result.case === INSTRUCTION) {
            setLogs(logs => [...logs, result.value.executableInstruction])
        }

        if (result.case === PROGRESS_INFO && (result.value.currentStepInfo[0] === EXECUTION_IN_PROGRESS || result.value.currentStepInfo[0] === STARTING_EXECUTION)) {
            const totalSteps = result.value.totalSteps
            const completedSteps = Math.max(result.value.currentStepNumber - 1, 0)
            const text = `Completed ${completedSteps} of ${result.value.totalSteps}`
            const progress = (completedSteps / totalSteps) * 100
            if (isNaN(progress) || progress < 0.0 || progress > 100.0) {
                setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_PROCESSING_INDETERMINATE, null))
            } else {
                setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_PROCESSING, progress))
            }
            setLogsExecutionStatusText(text)
        } else if (result.case === PROGRESS_INFO && result.value.currentStepInfo.length > 0) {
            setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_PROCESSING_INDETERMINATE))
        }

        if (result.case === PROGRESS_INFO && result.value.currentStepInfo.length > 0) {
            if (result.value.currentStepInfo[result.value.currentStepNumber] !== undefined) {
                setLogs(logs => [...logs, result.value.currentStepInfo[result.value.currentStepNumber]])
            }
        }

        if (result.case === INSTRUCTION_RESULT && result.value.serializedInstructionResult) {
            if (result.value.serializedInstructionResult.includes(SERVICE_IS_ADDED)) {
                getServices(enclave)
            }
            setLogs(logs => [...logs, result.value.serializedInstructionResult])
        }

        if (result.case === ERROR) {
            const errorMessage = result.value.error.value.errorMessage;
            setLogs(logs => [...logs, errorMessage])
        }

        if (result.case === RUN_FINISHED_EVENT) {
            console.log(result.value)
            if (result.value.isRunSuccessful) {
                setLogsExecutionStatusText("Script completed")
                setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_SUCCESS, null))
            } else {
                setLogsExecutionStatusText("Script failed")
                setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_ERROR, null))
            }
        }
    }

    useEffect(() => {
        setLoading(true)
        let stream;
        const fetchLogs = async () => {
            try {
                stream = await runStarlark(enclave.host, enclave.port, packageId, args, appData.jwtToken, appData.apiHost);
                for await (const res of stream) {
                    const result = res["runResponseLine"]
                    readStreamData(result)
                }
            } catch (ex) {
                console.error("Error occurred while reading data from the enclave: ", enclave.name)
                setLogsExecutionStatusText("Script failed to run with an error")
                setLogsCurrentExecutionStatus(logsStatus(SCRIPT_RUN_STATUS_ERROR, null))
            } finally {
                setLoading(false)
            }
        }
        fetchLogs();
    }, [packageId])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclave.name}/services/${service.serviceUuid}`, {state: {services, selected: service}})
    }

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

    useEffect(() => {
        setLogsComponent(
            (logs.length > 0) ?
                <Log
                    logs={logs}
                    fileName={enclave.name}
                    currentExecutionStatus={logsCurrentExecutionStatus}
                    executionStatusText={logsExecutionStatusText}
                /> : <Center color="white"> No Logs Available</Center>
        )
    }, [logs, logsCurrentExecutionStatus, logsExecutionStatusText])

    return (
        <div className="flex h-full">
            <div className="flex h-full">
                <LeftPanel
                    home={false}
                    heading={"Services"}
                    isServiceInfo={true}
                    renderList={() => renderServices(services, handleServiceClick)}
                />
                <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                    <div className='flex flex-col h-full space-y-1 bg-[#171923]'>
                        <Flex bg={"#171923"} height={`80px`}>
                            <Box p='2' m="4">
                                <Text color={"white"} fontSize='xl' as='b'>
                                    Logs for {enclave.name}
                                </Text>
                            </Box>
                            <Spacer/>
                        </Flex>
                        {(loading && logs.length === 0) ? <LoadingOverlay/> : logsComponent}
                    </div>
                </div>
                <RightPanel home={false} isServiceInfo={!loading} enclaveName={enclave.name}/>
            </div>
        </div>
    )
}

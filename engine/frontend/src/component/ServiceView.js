import {Badge, Box, Table, TableContainer, Tbody, Td, Tr} from "@chakra-ui/react";
import {CodeEditor} from "./CodeEditor";

const statusBadge = (status) => {
    let color = ""
    let text = ""
    if (status === "RUNNING" || status === 1) {
        color = "green"
        text = "RUNNING"
    } else if (status === "STOPPED" || status === 0) {
        color = "red"
        text = "STOPPED"
    } else if (status === "UNKNOWN" || status === 2) {
        color = "yellow"
        text = "UNKNOWN"
    } else {
        return (
            <Badge>N/A</Badge>
        )
    }
    return (
        <Badge colorScheme={color}>{text}</Badge>
    )
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
        <Tr key={heading} color="white">
            <Td><p><b>{heading}</b></p></Td>
            <Td>{displayData}</Td>
        </Tr>
    );
};

const codeBox = (serviceUuid, serviceName, parameterName, data) => {
    const serializedData = JSON.stringify(data, null, 2)
    const uniqueId = `${serviceUuid}-${serviceName}-${parameterName}.json`
    return (
        <Box>
            <CodeEditor
                uniqueId={uniqueId}
                readOnly={true}
                defaultWidthPx={250}
                defaultState={serializedData}
                autoFormat={true}
                showFormatButton={false}
                buttonSizes={"xs"}
            />
            </Box>
    );
}

const ServiceView = (service) => {
    const serviceInfo = service.service
    return (
        <TableContainer>
            <Table variant='simple' size='sm'>
                <Tbody>
                    {tableRow("Name", () => serviceInfo.name)}
                    {tableRow("UUID", () => <pre>{serviceInfo.serviceUuid}</pre>)}
                    {tableRow("Status", () => statusBadge(serviceInfo.container.status))}
                    {tableRow("Image", () => serviceInfo.container.imageName)}
                    {tableRow("Ports", () => codeBox(serviceInfo.shortenedUuid, serviceInfo.name, "ports", serviceInfo.ports, 0))}
                    {tableRow("ENTRYPOINT", () => codeBox(serviceInfo.shortenedUuid, serviceInfo.name, "entrypoint", serviceInfo.container.entrypointArgs, 1))}
                    {tableRow("CMD", () => codeBox(serviceInfo.shortenedUuid, serviceInfo.name, "cmd", serviceInfo.container.cmdArgs, 2))}
                    {tableRow("ENV", () => codeBox(serviceInfo.shortenedUuid, serviceInfo.name, "env", serviceInfo.container.envVars, 3))}
                </Tbody>
            </Table>
        </TableContainer>
    )
}

export default ServiceView;

import {Box, Button, Center, Checkbox, Flex, Grid, GridItem, Input, Stack, Text, Tooltip,} from '@chakra-ui/react'
import PackageCatalogOption from "./PackageCatalogOption";
import {useLocation, useNavigate} from "react-router";
import React, {useState} from 'react';
import startCase from 'lodash/startCase'
import {InfoOutlineIcon} from '@chakra-ui/icons'
import {CodeEditor} from "./CodeEditor";
import KeyValueTable from "./KeyValueTable";
import {useParams} from "react-router-dom";
import {getStarlarkRunConfig} from "../api/api";
import {useAppContext} from "../context/AppState";

const yaml = require("js-yaml")

const prettyPrintTypeSpecialCases = (type, arg) => {
    if (type === "LIST") {
        try {
            const dataType = getType(arg)
            if (dataType === undefined) return "LIST"
            return `${dataType} LIST`.toLowerCase()
        } catch (e) {
            return "LIST".toLowerCase()
        }
    }
    if (type === "DICT") {
        try {
            const x = getFirstSubType(arg)
            const y = getSecondSubType(arg)
            if (x === undefined || y === undefined) return "DICT"
            return `${x} -> ${y}`.toLowerCase()
        } catch (e) {
            return "DICT".toLowerCase()
        }
    }
    return type.toLowerCase()
}

const getArgName = (arg) => {
    return arg["name"]
}

const getType = (arg) => {
    return arg["typeV2"]["topLevelType"]
}

const getFirstSubType = (arg) => {
    return arg["typeV2"]["innerType1"]
}

const getSecondSubType = (arg) => {
    return arg["typeV2"]["innerType2"]
}

const isRequired = (arg) => {
    return arg["isRequired"]
}

const renderArgs = (args, handleChange, formData, errorData, packageName) => {
    return args.map((arg, index) => {

        // no need to process plan arg as it's internal!
        if (getArgName(arg) === "plan") {
            return
        }

        let dataType = "STRING";
        try {
            switch (getType(arg)) {
                case "INTEGER":
                    dataType = "INTEGER"
                    break;
                case "STRING":
                    dataType = "STRING"
                    break
                case "BOOL":
                    dataType = "BOOL"
                    break
                case "FLOAT":
                    dataType = "FLOAT"
                    break
                case "DICT":
                    dataType = "DICT"
                    break
                case "LIST":
                    dataType = "LIST"
                    break
                default:
                    dataType = "JSON"
            }
        } catch (e) {
            console.log("no data type provided, falling back to string")
        }

        return (
            <Flex color={"white"} key={`entry-${index}`}>
                <Flex mr="3" direction={"column"} w="15%">
                    <Text marginLeft={3}
                          align={"right"}
                          fontSize={"l"}
                    >
                        {startCase(arg.name)}
                        <Text
                            fontSize={"l"}
                            as="span"
                            color={"red"}
                        >
                            <Tooltip label="Required variable">
                                <span>{arg.isRequired ? " *" : null}</span>
                            </Tooltip>

                        </Text>
                        <Tooltip label={arg.description}>
                            <InfoOutlineIcon marginLeft={2}/>
                        </Tooltip>

                    </Text>
                    <Text marginLeft={3} as='kbd' fontSize='xs'
                          align={"right"}>{prettyPrintTypeSpecialCases(dataType, arg)}</Text>
                </Flex>
                <Flex flex="1" mr="3" direction={"column"}>
                    {errorData[index].length > 0 ?
                        <Text marginLeft={3} align={"left"} fontSize={"xs"}
                              color="red.500">
                            {errorData[index]}
                        </Text> : null
                    }
                    {renderSingleArg(arg.name, dataType, errorData, formData, index, handleChange, packageName)}
                </Flex>
            </Flex>
        )
    })
}

const renderSingleArg = (fieldName, type, errorData, formData, index, handleChange, packageName) => {
    const uniqueId = `${packageName}-${fieldName}`
    switch (type) {
        case "INTEGER":
        case "STRING":
        case "BOOL":
        case "LIST":
        case "FLOAT":
            return (
                <Input
                    color='gray.300'
                    onChange={e => handleChange(e.target.value, index)}
                    value={formData[index]}
                    borderColor={errorData[index] ? "red.400" : null}
                />
            )

        case "JSON":
            return (
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >
                    {CodeEditor(
                        uniqueId,
                        (data) => handleChange(data, index)
                    )}
                </Box>
            )
        case "DICT":
            return (
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >
                    {KeyValueTable((data) => handleChange(data, index))}
                </Box>
            )

        default:
            return <p key={`data-${index}`}>Unsupported data type encountered</p>
    }
}


const checkValidUndefinedType = (data) => {
    try {
        const val = yaml.load(data)
    } catch (ex) {
        return false;
    }
    return true;
}

const checkValidJsonType = (data) => {
    if (data === undefined || data === "undefined" || data.length === 0) {
        return false
    }

    try {
        JSON.parse(data)
        return true;
    } catch (ex) {
        return false
    }
}

const checkValidStringType = (data) => {
    if (data === "undefined" || data.length === 0) {
        return false
    }

    try {
        const val = JSON.parse(data)
        if (typeof val === "string") {
            return true
        }
        return false
    } catch (ex) {
        if (data.includes("\"") || data.includes("\'")) {
            return false
        }
        return true;
    }
}

const checkValidIntType = (data) => {
    const isNumeric = (value) => {
        return /^-?\d+$/.test(value);
    }
    if (data === "undefined") {
        return false
    }
    try {
        const trimmedData = data.trim()
        return isNumeric(trimmedData)
    } catch (ex) {
        return false
    }
}

const checkValidFloatType = (data) => {
    const isValidFloat = (value) => {
        return !isNaN(Number(value))
    }
    if (data === "undefined") {
        return false
    }

    const trimmedData = data.trim()
    if (trimmedData.length === 0) {
        return false
    }

    return isValidFloat(trimmedData)
}

const checkValidBooleanType = (data) => {
    const trimData = data.trim()
    return ["TRUE", "FALSE"].includes(trimData.toUpperCase())
}

const checkValidListType = (data, subType) => {
    if (data === "undefined") {
        return false
    }
    try {
        parseList(data, subType)
        return true
    } catch (ex) {
        return false
    }
}


// Referencing this list:
// https://github.com/kurtosis-tech/kurtosis-package-indexer/blob/main/server/crawler/starlark_value_type.go#L25-L38
// 		0: "BOOL",
// 		1: "STRING",
// 		2: "INTEGER",
// 		3: "FLOAT",
// 		4: "DICT",
// 		5: "JSON",
// 		6: "LIST",
// 	}
const convertDataTypeToJsType = (dataType) => {
    const normalizedDataType = dataType.trim().toLowerCase()
    switch (normalizedDataType) {
        case "int":
        case "integer":
        case "float":
            return "number"
        case "str": // Added python type for good measure
        case "string":
            return "string"
        case "bool":
        case "boolean":
            return "boolean"
        case "dict":
        case "json":
            return "object"
    }
}

const parseList = (data, rawDataType) => {
    const dataType = rawDataType
    const convertedDataType = convertDataTypeToJsType(dataType)
    const trimData = data.trim()
    const jsonData = `[${trimData}]`
    const parsedJson = JSON.parse(jsonData)
    parsedJson.forEach(item => {
        const type = typeof item;
        if (type !== convertedDataType) {
            throw `Invalid datatype: Expected ${convertedDataType} but got ${type}`;
        }
    })
    return parsedJson
}


const loadPackageRunConfig = async (host, port, token, apiHost) => {
    const data = await getStarlarkRunConfig(host, port, token, apiHost)
    // consoloe.log(data)
    return data;
}

const PackageCatalogForm = ({createEnclave, mode}) => {
    const {appData} = useAppContext()
    const navigate = useNavigate()
    const location = useLocation()
    const {state} = location;
    const [runningPackage, setRunningPackage] = useState(false)
    const [enclaveName, setEnclaveName] = useState("")
    const [productionMode, setProductionMode] = useState(false)

    let thisKurtosisPackage = {};
    let initialFormData = {}
    let initialErrorData = {}

    console.log(state)

    if (mode === "create") {
        const {kurtosisPackage} = state
        thisKurtosisPackage = kurtosisPackage
    } else if (mode === "edit") {
        const {name, host, port} = state
        const runConfigPromise = loadPackageRunConfig(host, port, appData.jwtToken, appData.apiHost)
        runConfigPromise.then((runConfig) =>{
            console.log(runConfig)
        })


    } else {
        console.log("Unsupported package configuration mode", mode)
    }

    console.log("thisKurtosisPackage", thisKurtosisPackage)
    console.log("thisKurtosisPackage", JSON.stringify(thisKurtosisPackage))

    thisKurtosisPackage?.args.forEach(
        (arg, index) => {
            if (arg.name !== "plan") {
                initialFormData[index] = ""
            }
        }
    )

    const [formData, setFormData] = useState(initialFormData)

    thisKurtosisPackage.args.forEach((arg, index) => {
        if (arg.name !== "plan") {
            initialErrorData[index] = ""
        }
    })
    const [errorData, setErrorData] = useState(initialErrorData)

    const handleFormDataChange = (value, index) => {
        const newData = {
            ...formData,
            [index]: value,
        }
        setFormData(newData)

        if (errorData[index]) {
            const newErrorData = {
                ...errorData,
                [index]: false,
            }
            setErrorData(newErrorData)
        }
    }

    const handleCancelBtn = () => {
        navigate("/catalog")
    }

    const handleRunBtn = () => {
        let errorsFound = {}

        Object.keys(formData).filter(key => {
            const arg = thisKurtosisPackage.args[key]
            let type = ""
            try {
                type = getType(arg)
            } catch {
            }
            const required = isRequired(arg)

            // if it's optional and empty it's fine
            if (!required && formData[key].length === 0) {
                return
            }

            let valid = true
            if (type === "STRING") {
                valid = checkValidStringType(formData[key])
            } else if (type === "INTEGER") {
                valid = checkValidIntType(formData[key])
            } else if (type === "BOOL") {
                valid = checkValidBooleanType(formData[key])
            } else if (type === "FLOAT") {
                valid = checkValidFloatType(formData[key])
            } else if (type === "LIST") {
                let subType = getFirstSubType(arg)
                valid = checkValidListType(formData[key], subType)
            } else if (type === "DICT") {
                valid = checkValidJsonType(formData[key])
            } else if (type === "JSON") {
                valid = checkValidJsonType(formData[key])
            } else {
                valid = checkValidUndefinedType(formData[key])
            }

            let typeToPrint = type
            if (type === undefined) {
                typeToPrint = "JSON"
            } else if (type === "BOOL") {
                typeToPrint = "BOOLEAN (TRUE/FALSE)"
            } else {
                typeToPrint = prettyPrintTypeSpecialCases(type, thisKurtosisPackage.args[key])
            }

            if (!valid) {
                errorsFound[key] = `Incorrect type: expected ${typeToPrint}`;
            }
        })

        Object.keys(formData).filter(key => {
            const required = isRequired(thisKurtosisPackage.args[key])
            let valid = true;
            if (required) {
                if (formData[key].length === 0) {
                    valid = false;
                }
            }

            if (!valid) {
                errorsFound[key] = `This field is required and cannot be empty`;
            }
        })

        if (Object.keys(errorsFound).length === 0) {
            setRunningPackage(true)
            let args = {}
            Object.keys(formData).map(key => {
                const arg = thisKurtosisPackage.args[key]
                const argName = getArgName(arg)
                let type = ""
                try {
                    type = getType(arg)
                } catch {
                }
                const value = formData[key]

                let val;
                if (value.length > 0) {
                    if (type === "INTEGER") {
                        val = parseInt(value)
                        args[argName] = val
                    } else if (type === "BOOL") {
                        val = value.toUpperCase()
                        args[argName] = (val === "TRUE") ? true : false
                    } else if (type === "FLOAT") {
                        val = parseFloat(value)
                        args[argName] = val
                    } else if (type === "LIST") {
                        let subType = getFirstSubType(thisKurtosisPackage, key)
                        val = parseList(value, subType)
                        args[argName] = val
                    } else if (type === "STRING") {
                        args[argName] = value
                    } else {
                        val = JSON.parse(value)
                        args[argName] = val
                    }
                }
            })

            const stringifiedArgs = JSON.stringify(args)
            const runKurtosisPackageArgs = {
                packageId: thisKurtosisPackage.name,
                args: stringifiedArgs,
            }

            handleCreateEnclave(runKurtosisPackageArgs, enclaveName, productionMode)

        } else {
            const newErrorData = {
                ...errorData,
                ...errorsFound
            }
            setErrorData(newErrorData)
        }
    }

    const handleCreateEnclave = async (runKurtosisPackageArgs, enclaveName, productionMode) => {
        await createEnclave(runKurtosisPackageArgs, enclaveName, productionMode)
        setRunningPackage(false)
    }

    return (
        <div className='w-screen'>
            <Grid
                templateAreas={`"option"
                                "packageId"
                                "main"
                                "configure"`}
                gridTemplateRows={'60px 60px 1fr 60px'}
                gridTemplateColumns={'1fr'}
                h='100%'
                w='100%'
                color='blackAlpha.700'
                fontWeight='bold'
                gap={2}
            >
                <GridItem area={'option'} pt="1">
                    <PackageCatalogOption catalog={true}/>
                </GridItem>
                <GridItem area={'packageId'} p="1">
                    <Flex direction={"column"} gap={"2"}>
                        <Center>
                            <Text color={"white"} fontSize={"2xl"}> {thisKurtosisPackage.name} </Text>
                        </Center>
                        <Checkbox
                            marginLeft={2}
                            color={"white"}
                            fontSize={"xl"}
                            isChecked={productionMode}
                            onChange={(e) => setProductionMode(e.target.checked)}
                        >
                            <Text>
                                Restart services
                                <Tooltip
                                    label="When enabled, Kurtosis will automatically restart any services that crash inside the enclave">
                                    <InfoOutlineIcon marginLeft={2}/>
                                </Tooltip>

                            </Text>
                        </Checkbox>
                    </Flex>
                </GridItem>
                <GridItem area={'main'} h="90%" overflowY={"scroll"} mt="10">
                    <Stack spacing={4}>
                        <Flex color={"white"}>
                            <Flex mr="3" direction={"column"} w="15%">
                                <Text
                                    marginLeft={3}
                                    align={"right"}
                                    fontSize={"l"}
                                >Enclave Name
                                    <Tooltip label="Leave empty to auto-generate an enclave name">
                                        <InfoOutlineIcon marginLeft={2}/>
                                    </Tooltip>
                                </Text>
                            </Flex>
                            <Flex flex="1" mr="3" direction={"column"}>
                                <Input
                                    color='gray.300'
                                    value={enclaveName}
                                    onChange={(e) => setEnclaveName(e.target.value)}
                                />
                            </Flex>
                        </Flex>
                        {renderArgs(thisKurtosisPackage.args, handleFormDataChange, formData, errorData, thisKurtosisPackage.name)}
                    </Stack>
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Flex gap={5}>
                        <Button colorScheme='red' w="50%" onClick={handleCancelBtn}> Cancel </Button>
                        <Button bg='#24BA27'
                                w="50%"
                                onClick={handleRunBtn}
                                isLoading={runningPackage}
                                loadingText="Running..."
                        >
                            Run
                        </Button>
                    </Flex>
                </GridItem>
            </Grid>
        </div>
    );
};
export default PackageCatalogForm;

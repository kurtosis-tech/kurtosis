import {Box, Button, Center, Checkbox, Flex, Grid, GridItem, Input, Stack, Text, Tooltip,} from '@chakra-ui/react'
import PackageCatalogOption from "./PackageCatalogOption";
import {useLocation, useNavigate} from "react-router";
import React, {useEffect, useState} from 'react';
import startCase from 'lodash/startCase'
import {InfoOutlineIcon} from '@chakra-ui/icons'
import {CodeEditor} from "./CodeEditor";
import KeyValueTable from "./KeyValueTable";
import {getStarlarkRunConfig} from "../api/api";
import {useAppContext} from "../context/AppState";
import {getKurtosisPackages} from "../api/packageCatalog";

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

        let dataType = "JSON";
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
                    {errorData && errorData[index] && errorData[index].length > 0 ?
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
    const dataCallback = (data) => handleChange(data, index)
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
            const processedState = formData[index] === "" ? "{}" : formData[index]
            return (
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >
                    <CodeEditor
                        uniqueId={uniqueId}
                        dataCallback={dataCallback}
                        defaultState={processedState}
                    />
                </Box>
            )
        case "DICT":
            return (
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >
                    <KeyValueTable
                        dataCallback={dataCallback}
                        defaultState={formData[index]}
                    />
                </Box>
            )

        default:
            return <p key={`data-${index}`}>Unsupported data type encountered</p>
    }
}


const checkValidUndefinedType = (data) => {
    try {
        yaml.load(data)
    } catch (ex) {
        return false;
    }
    return true;
}

const checkValidJsonType = (rawData, required) => {
    // data is normally going to be a string serialized json object as it comes from the CodeEditor, so we try and process it as such
    if (typeof rawData === "string") {
        try {
            const parsed = JSON.parse(rawData)
            if (required && Object.keys(parsed).length < 1) {
                return false
            }
            return true;
        } catch (ex) {
            console.error("Data is not serialized json", rawData)
            return false
        }
    } else if (typeof rawData === "object") {
        // if it's already an object then we only check that it's non-empty (if it's required)
        try {
            if (required && Object.keys(rawData).length < 1) {
                return false
            }
            return true;
        } catch (ex) {
            return false
        }
    }
    console.error(`Data is unknown type ${typeof rawData}`, rawData)
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
        const [thisKurtosisPackage, setThisKurtosisPackage] = useState({})
        const [formData, setFormData] = useState({})
        const [errorData, setErrorData] = useState({})

        const isEditMode = () => {
            return mode === "edit"
        }

        const serializeIfJson = (data) => {
            let processedData = data // assumption is it's not json
            try {
                // serialize if it's json object
                if (typeof processedData === 'object') {
                    processedData = JSON.stringify(processedData)
                }
            } catch {
            }
            return processedData
        }

        const updateThisPackage = (currentPackage, formData, errorData, existingParamsMap) => {
            let initialFormData = {}
            let initialErrorData = {}
            currentPackage.args.forEach(
                (arg, index) => {
                    if (arg.name !== "plan") {
                        initialFormData[index] = ""
                    }
                    if (existingParamsMap && existingParamsMap[arg.name]) {
                        initialFormData[index] = serializeIfJson(existingParamsMap[arg.name])
                    }
                }
            )
            currentPackage.args.forEach((arg, index) => {
                if (arg.name !== "plan") {
                    initialErrorData[index] = ""
                }
            })
            setFormData(initialFormData)
            setErrorData(initialErrorData)
            setThisKurtosisPackage(currentPackage)
        }

        useEffect(() => {
            if (mode === "create") {
                const {kurtosisPackage} = state
                updateThisPackage(kurtosisPackage, formData, errorData, null)
            } else if (isEditMode()) {
                const {name, host, port, enclave} = state
                getKurtosisPackages() // TODO: Debounce this call. Only call once!
                    .then((packages) => {
                        loadPackageRunConfig(host, port, appData.jwtToken, appData.apiHost)
                            .then((runConfig) => {
                                const matchedPackage = packages.find((p) => p.name === runConfig.packageId)
                                setEnclaveName(name)
                                const existingParamsMap = JSON.parse(runConfig.serializedParams || '{}')
                                if (matchedPackage) {
                                    updateThisPackage(matchedPackage, formData, errorData, existingParamsMap)
                                } else {
                                    console.error(`Was not able to match a package to the running enclave ${runConfig.name} for package id: ${runConfig.packageId}`)
                                }
                            })
                    })
            } else {
                console.error(`Unsupported package configuration mode: ${mode}`)
            }
        }, [])

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
                    valid = checkValidJsonType(formData[key], required)
                } else if (type === "JSON") {
                    // required = false, always because we have a later check that ensures the object is not null
                    valid = checkValidJsonType(formData[key], false)
                } else {
                    valid = checkValidUndefinedType(formData[key])
                }

                let typeToPrint
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
                            val = value.trim().toUpperCase()
                            args[argName] = val === "TRUE"
                        } else if (type === "FLOAT") {
                            val = parseFloat(value)
                            args[argName] = val
                        } else if (type === "LIST") {
                            let subType = getFirstSubType(thisKurtosisPackage, key)
                            val = parseList(value, subType)
                            args[argName] = val
                        } else if (type === "STRING") {
                            args[argName] = value
                        } else if (type === "DICT") {
                            if (typeof value === "string") {
                                try {
                                    val = JSON.parse(value)
                                    // Check that the object is non-empty before adding it
                                    if (Object.keys(val).length > 0) {
                                        args[argName] = val
                                    }
                                } catch {
                                    console.error(`Expected data to be serialized json, but parsing failed. arg=${argName}, value=${value}`)
                                }
                            } else if (typeof value === "object") {
                                // Check that the object is non-empty before adding it
                                if (Object.keys(value).length > 0) {
                                    args[argName] = value
                                }
                            } else {
                                console.error(`Data field '${argName}' was not a valid object but was type ${typeof value}. Contained value: '${value}'`)
                            }
                        } else {
                            try {
                                val = JSON.parse(value)
                                args[argName] = val
                            } catch (ex) {
                                console.error(`Data field '${argName}' was not a valid object but was type ${typeof value}. Contained value: '${value}'`)
                            }
                        }
                    }
                })

                console.log("raw args", args)
                const stringifiedArgs = JSON.stringify(args)
                const runKurtosisPackageArgs = {
                    packageId: thisKurtosisPackage.name,
                    args: stringifiedArgs,
                }
                console.log("Creating enclave with package args", runKurtosisPackageArgs)

                if (!isEditMode()) {
                    handleCreateEnclave(runKurtosisPackageArgs, enclaveName, productionMode, mode, null)
                } else {
                    const {enclave} = state
                    handleCreateEnclave(runKurtosisPackageArgs, enclaveName, productionMode, mode, enclave)
                }

            } else {
                const newErrorData = {
                    ...errorData,
                    ...errorsFound
                }
                setErrorData(newErrorData)
            }
        }

        const handleCreateEnclave = async (runKurtosisPackageArgs, enclaveName, productionMode, mode, maybeExistingEnclave) => {
            await createEnclave(runKurtosisPackageArgs, enclaveName, productionMode, mode, maybeExistingEnclave)
            setRunningPackage(false)
        }

        const [renderedArgs, setRenderedArgs] = useState(<></>)

        useEffect(() => {
                if (thisKurtosisPackage.args && formData && errorData) {
                    setRenderedArgs(
                        renderArgs(thisKurtosisPackage.args, handleFormDataChange, formData, errorData, thisKurtosisPackage.name)
                    )
                }
            }, [thisKurtosisPackage.args, formData, errorData, thisKurtosisPackage.name]
        )

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
                                        disabled={isEditMode()}
                                    />
                                </Flex>
                            </Flex>
                            {renderedArgs}
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
    }
;
export default PackageCatalogForm;

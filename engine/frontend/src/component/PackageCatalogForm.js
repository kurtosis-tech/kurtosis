import {
    Box,
    Button,
    Center,
    Checkbox,
    Flex,
    Grid,
    GridItem, HStack,
    Input, InputGroup, InputLeftAddon, Spacer,
    Stack,
    Text,
    Textarea,
    Tooltip,
} from '@chakra-ui/react'
import PackageCatalogOption from "./PackageCatalogOption";
import {useLocation, useNavigate} from "react-router";
import React, {useEffect, useState} from 'react';
import startCase from 'lodash/startCase'
import {InfoOutlineIcon} from '@chakra-ui/icons'
import {ObjectInput} from 'react-object-input'


const yaml = require("js-yaml")
const KeyValueTable = (dataCallBack) => {
    const [value, setValue] = useState({})

    useEffect(() => {
        dataCallBack(JSON.stringify(value))
    }, [value])

    return (
        <Box
            border="1px"
            borderColor='gray.700'
            borderRadius="7"
            margin={"1px"}
            padding={1}
        >
            <ObjectInput
                obj={value}
                onChange={setValue}
                renderItem={(key, value, updateKey, updateValue, deleteProperty) => (
                    <Box
                        margin={1}
                    >
                        <HStack
                            spacing={1}
                            direction="row"
                        >
                            <InputGroup>
                                <InputLeftAddon children='Key'/>
                                <Input
                                    type="text"
                                    value={key}
                                    onChange={e => updateKey(e.target.value)}
                                    size="md"
                                    variant='filled'
                                    // htmlSize={10} width='auto'
                                />
                            </InputGroup>

                            <InputGroup>
                                <InputLeftAddon children='Value'/>
                                <Input
                                    type="text"
                                    value={value || ''} // value will be undefined for new rows
                                    onChange={e => updateValue(e.target.value)}
                                    size="md"
                                    variant='filled'
                                    // htmlSize={10} width='auto'
                                />
                            </InputGroup>
                            <Button
                                // margin={"10px"}
                                onClick={deleteProperty}
                            >
                                x
                            </Button>
                        </HStack>
                    </Box>
                )}
                renderAdd={addItem => <Button margin={1} onClick={addItem}>Add item</Button>}
                // renderEmpty={() => <p></p>}
            />
        </Box>
    )
}
const renderArgs = (args, handleChange, formData, errorData) => {
    return args.map((arg, index) => {
        let placeholder = "";
        switch (arg.type) {
            case "INTEGER":
                placeholder = "INTEGER"
                break;
            case "STRING":
                placeholder = "STRING"
                break
            case "BOOL":
                placeholder = "BOOL"
                break
            case "FLOAT":
                placeholder = "FLOAT"
                break
            case "KEY_VALUE":
                placeholder = "KEY_VALUE"
                break
            default:
                placeholder = "JSON"
        }

        // no need to show plan arg as it's internal!
        if (arg.name === "plan") {
            return
        }

        // REMOVE: JUST FOR TESTING
        if(arg.name === "remote_chains"){
            placeholder="KEY_VALUE"
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
                        <Tooltip label="Some text coming here soon">
                            <InfoOutlineIcon marginLeft={2}/>
                        </Tooltip>

                    </Text>
                    <Text marginLeft={3} as='kbd' fontSize='xs' align={"right"}>{placeholder.toLowerCase()}</Text>
                </Flex>
                <Flex flex="1" mr="3" direction={"column"}>
                    {errorData[index].length > 0 ?
                        <Text marginLeft={3} align={"left"} fontSize={"xs"}
                              color="red.500"> {errorData[index]} </Text> : null}
                    {renderSingleArg(placeholder, errorData, formData, index, handleChange)}
                </Flex>
            </Flex>
        )
    })
}

const renderSingleArg = (type, errorData, formData, index, handleChange) => {
    switch (type) {
        case "INTEGER":
        case "STRING":
        case "BOOL":
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
                <Textarea
                    borderColor={errorData[index] ? "red.400" : null}
                    minHeight={"200px"}
                    onChange={e => handleChange(e.target.value, index)}
                    value={formData[index]}
                />
            )
        case "KEY_VALUE":
            return KeyValueTable((data) => handleChange(data, index))

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

const PackageCatalogForm = ({handleCreateNewEnclave}) => {
    const navigate = useNavigate()
    const location = useLocation()
    const {state} = location;
    const {kurtosisPackage} = state
    const [runningPackage, setRunningPackage] = useState(false)
    const [enclaveName, setEnclaveName] = useState("")
    const [productionMode, setProductionMode] = useState(false)

    let initialFormData = {}
    kurtosisPackage.args.forEach(
        (arg, index) => {
            if (arg.name !== "plan") {
                initialFormData[index] = ""
            }
        }
    )
    const [formData, setFormData] = useState(initialFormData)

    let initialErrorData = {}
    kurtosisPackage.args.map((arg, index) => {
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
            let type = kurtosisPackage.args[key]["type"]
            const required = kurtosisPackage.args[key]["isRequired"]

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
            } else {
                valid = checkValidUndefinedType(formData[key])
            }

            let typeToPrint = type
            if (type === undefined) {
                typeToPrint = "JSON"
            }

            if (type === "BOOL") {
                typeToPrint = "BOOLEAN (TRUE/FALSE)"
            }

            if (!valid) {
                errorsFound[key] = `Incorrect type, expected ${typeToPrint}`;
            }
        })

        Object.keys(formData).filter(key => {
            let valid = true;
            const required = kurtosisPackage.args[key]["isRequired"]
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
                const argName = kurtosisPackage.args[key].name
                const value = formData[key]

                let val;
                let trimmedValue;
                if (value.length > 0) {
                    if (kurtosisPackage.args[key]["type"] === "INTEGER") {
                        trimmedValue = value.trim()
                        val = parseInt(value)
                        args[argName] = val
                    } else if (kurtosisPackage.args[key]["type"] === "BOOL") {
                        trimmedValue = value.trim()
                        val = value.toUpperCase()
                        args[argName] = (val === "TRUE") ? true : false
                    } else if (kurtosisPackage.args[key]["type"] === "FLOAT") {
                        trimmedValue = value.trim()
                        val = parseFloat(value)
                        args[argName] = val
                    } else if (kurtosisPackage.args[key]["type"] === "STRING") {
                        args[argName] = value
                    } else {
                        val = JSON.parse(value)
                        args[argName] = val
                    }
                }
            })

            const stringifiedArgs = JSON.stringify(args)
            const runKurtosisPackageArgs = {
                packageId: kurtosisPackage.name,
                args: stringifiedArgs,
            }

            handleCreateNewEnclave(runKurtosisPackageArgs, enclaveName, productionMode)

        } else {
            const newErrorData = {
                ...errorData,
                ...errorsFound
            }
            setErrorData(newErrorData)
        }
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
                            <Text color={"white"} fontSize={"2xl"}> {kurtosisPackage.name} </Text>
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
                                    label="When enabled, Kurtosis will automatically restart any services that crash in side the enclave">
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
                        {renderArgs(kurtosisPackage.args, handleFormDataChange, formData, errorData)}
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






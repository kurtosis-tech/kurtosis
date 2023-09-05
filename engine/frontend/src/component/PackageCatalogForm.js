import { 
    Grid, 
    GridItem, 
    Center,  
    Input, 
    Flex, 
    Button,
    Stack,
    Text,
    Textarea,
    Checkbox,
} from '@chakra-ui/react'
import PackageCatalogOption from "./PackageCatalogOption";
import { useLocation, useNavigate } from "react-router";
import {useState} from 'react';

const yaml = require("js-yaml")

const renderArgs = (args, handleChange, formData, errorData) => {
    return args.map((arg, index) => {
        let placeholder = "";
        switch(arg.type) {
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
            default:
                placeholder = "JSON"
        }

        // no need to show plan arg as it's internal!
        if (arg.name === "plan") {
            return 
        }
        
        return (
            <Flex color={"white"}>
                <Flex mr="3" direction={"column"} w="15%">
                    <Text align={"center"} fontSize={"xl"}> {arg.name} </Text>
                    {arg.isRequired ? <Text align={"center"} fontSize={"s"} color="red.500"> Required</Text>: null}
                </Flex>
                <Flex flex="1" mr="3" direction={"column"}>
                {   errorData[index].length > 0 ? <Text align={"center"} fontSize={"s"} color="red.500"> {errorData[index]} </Text> : null}
                    {
                        ["INTEGER", "STRING", "BOOL", "FLOAT"].includes(placeholder) ? <Input placeholder={placeholder} color='gray.300' onChange={e => handleChange(e.target.value, index)} value={formData[index]} borderColor={errorData[index] ? "red.400": null}/> :
                        <Textarea borderColor={errorData[index] ? "red.400": null} placeholder={placeholder} minHeight={"200px"} onChange={e => handleChange(e.target.value, index)} value={formData[index]}/>
                    }   
                </Flex>
            </Flex>
        )
    })
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
        return isNumeric(data)
    } catch(ex) {
        return false
    }
}

const checkValidFloatType = (data) => {
    const  isValidFloat = (value) => {
        return (/^-?[\d]*(\.[\d]+)?$/g).test(value);
    }
    if (data === "undefined") {
        return false
    }
    return isValidFloat(data)
}

const checkValidBooleanType = (data) => {
    return ["TRUE", "FALSE"].contains(data)
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
    kurtosisPackage.args.map(
        (arg, index)=> {
            if (arg.name !== "plan") {
                initialFormData[index] = ""
            }
        }
     )
    const [formData, setFormData] = useState(initialFormData)

    let initialErrorData = {}
    kurtosisPackage.args.map((arg, index)=> {
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
                errorsFound[key] = `Incorrect type, expected ${type}`;
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
                if (value.length > 0) {
                    if (kurtosisPackage.args[key]["type"] === "INTEGER") {
                        val = parseInt(value)
                        args[argName] = val
                    } else if (kurtosisPackage.args[key]["type"] === "BOOL") {
                        args[argName] = (value === "TRUE") ? true : false
                    } else if (kurtosisPackage.args[key]["type"] === "FLOAT") {
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
                    <PackageCatalogOption catalog={true} />
                </GridItem>
                <GridItem area={'packageId'} p="1">
                    <Flex direction={"column"} gap={"2"}>
                        <Center>
                            <Text color={"white"} fontSize={"4xl"}> {kurtosisPackage.name} </Text>
                         </Center>
                         <Center>
                            <Checkbox color={"white"} fontSize={"2xl"} isChecked={productionMode} onChange={(e)=>setProductionMode(e.target.checked)}> 
                                <Text color={"white"} fontSize={"xl"} textAlign={"justify-center"}> 
                                    Production Mode 
                                </Text>
                            </Checkbox>
                        </Center>
                    </Flex>
                </GridItem>
                <GridItem area={'main'} h="90%" overflowY={"scroll"} mt="10"> 
                    <Stack spacing={4}>
                        <Flex color={"white"}>
                            <Flex mr="3" direction={"column"} w="15%">
                                <Text align={"center"} fontSize={"xl"}> Enclave Name </Text>
                            </Flex>
                            <Flex flex="1" mr="3" direction={"column"}>
                                <Input 
                                    placeholder={"IF NOT PROVIDED, THIS WILL BE GENERATED AUTOMATICALLY"} 
                                    color='gray.300' 
                                    value={enclaveName}
                                    onChange={(e)=>setEnclaveName(e.target.value)}
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






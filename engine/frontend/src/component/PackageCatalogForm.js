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

const PackageCatalogForm = ({handleCreateNewEnclave}) => {
    const navigate = useNavigate()
    const location = useLocation()
    const {state} = location;
    const {kurtosisPackage} = state

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
            } else {
                valid = checkValidUndefinedType(formData[key])
            }

            if (type === undefined) {
                type = "JSON"
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
            let args = {}
            Object.keys(formData).map(key => {
                const argName = kurtosisPackage.args[key].name
                const value = formData[key]
                if (!["INTEGER", "STRING", "BOOL", "FLOAT"].includes(kurtosisPackage.args[key]["type"])) {
                    try {
                        const val = JSON.parse(value)
                        console.log(val)
                        args[argName] = val
                    } catch(ex) {
                        console.log("this error should not come up")
                    }
                } else {
                    args[argName] = value
                }
            })

            const stringifiedArgs = JSON.stringify(args)
            console.log("stream ", stringifiedArgs)
            const runKurtosisPackageArgs = {
                packageId: kurtosisPackage.name,
                args: stringifiedArgs,
            }
            handleCreateNewEnclave(runKurtosisPackageArgs)

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
                    <PackageCatalogOption />
                </GridItem>
                <GridItem area={'packageId'} p="1">
                    <Center>
                        <Text color={"white"} fontSize={"4xl"}> {kurtosisPackage.name} </Text>
                    </Center>
                </GridItem>
                <GridItem area={'main'} h="90%" overflowY={"scroll"} mt="10"> 
                    <Stack spacing={4}>
                        {renderArgs(kurtosisPackage.args, handleFormDataChange, formData, errorData)}
                    </Stack>
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Flex gap={5}>
                        <Button colorScheme='red' w="50%" onClick={handleCancelBtn}> Cancel </Button>
                        <Button bg='#24BA27' w="50%" onClick={handleRunBtn}> Run </Button>
                    </Flex>
                </GridItem>
            </Grid>
        </div>
    );
  };
  export default PackageCatalogForm;






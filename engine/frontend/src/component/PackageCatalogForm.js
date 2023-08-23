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
import { useLocation } from "react-router";
import {useState} from 'react';


const renderArgs = (args, handleChange, formData, errorData) => {
    return args.map((arg, index) => {
        let placeholder = "";
        switch(arg.type) {
            case "int":
              placeholder = "int"
              break;
            case "string":
                placeholder = "string"
                break
            default:
                placeholder = "YAML/JSON"
        }
        
        return (
            <Flex color={"white"}>
                <Flex w="15%" mr="3" direction={"column"}>
                    <Text align={"center"} fontSize={"2xl"}> {arg.name} </Text>
                    {arg.required ? <Text align={"center"} fontSize={"s"} color="red.500"> Required</Text>: null}
                </Flex>
                <Flex flex="1" mr="3" direction={"column"}>
                {   errorData[index] ? <Text align={"center"} fontSize={"s"} color="red.500"> Incorrect type, expected {placeholder} </Text> : null}
                    {
                        ["int", "string"].includes(placeholder) ? <Input placeholder={placeholder} color='gray.300' onChange={e => handleChange(e.target.value, index)} value={formData[index]} borderColor={errorData[index] ? "red.400": null}/> :
                        <Textarea borderColor={errorData[index] ? "red.400": null} placeholder={placeholder} minHeight={"200px"} onChange={e => handleChange(e.target.value, index)} value={formData[index]}/>
                    }   
                </Flex>
            </Flex>
        )
    })
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
            console.log(data)
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

const PackageCatalogForm = () => {

    const location = useLocation()
    const {state} = location;
    const {kurtosisPackage} = state

    let initialFormData = {}
    kurtosisPackage.args.map((arg, index)=> initialFormData[index] = "")
    const [formData, setFormData] = useState(initialFormData)

    let initialErrorData = {}
    kurtosisPackage.args.map((arg, index)=> initialErrorData[index] = false)
    const [errorData, setErrorData] = useState(initialErrorData)
    
    console.log(errorData)

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

    const handleRunBtn = () => {
        let errorFound = false;
        let errorsFound = {}

        Object.keys(formData).filter(key => {
            const type = kurtosisPackage.args[key]["type"]
            let valid = true

            if (type === "string") {
                valid = checkValidStringType(formData[key])
                console.log(valid)
            } else if (type === "int") {
                valid = checkValidIntType(formData[key])
            }

            if (!valid) {
                errorsFound[key] = true;
            }
        })

        if (Object.keys(errorsFound).length === 0) {
            console.log("success")
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
                <GridItem area={'main'} h="100%" overflowY={"scroll"} mt="10"> 
                    <Stack spacing={4}>
                        {renderArgs(kurtosisPackage.args, handleFormDataChange, formData, errorData)}
                    </Stack>
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Flex gap={5}>
                        <Button colorScheme='red' w="50%" > Cancel </Button>
                        <Button colorScheme='green' w="50%" onClick={handleRunBtn}> Run </Button>
                    </Flex>
                </GridItem>
            </Grid>
        </div>
    );
  };
  export default PackageCatalogForm;






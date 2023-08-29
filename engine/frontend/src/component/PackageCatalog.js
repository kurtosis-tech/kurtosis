import {useEffect, useState} from "react";
import { 
    Grid, 
    GridItem, 
    Center, 
    List, 
    ListItem, 
    InputGroup, 
    InputLeftElement, 
    Input, 
    Button,
    Text
} from '@chakra-ui/react'

import { SearchIcon } from '@chakra-ui/icons'
import PackageCatalogOption from "./PackageCatalogOption";
import { useNavigate } from "react-router";

const PackageCatalog = ({kurtosisPackages: defaultPackages}) => {

    const [kurtosisPackages, setKurtosisPackages] = useState([])
    const navigate = useNavigate()
    const [chosenPackage, setChosenPackage] = useState(null)

    useEffect(()=> {
        setKurtosisPackages(defaultPackages)
    }, [defaultPackages.length])

    const selectPackage = (index) => {
        if (index === chosenPackage) {
            setChosenPackage(null)
        } else {
            setChosenPackage(index)
        }
    }

    const handleConfigureButtonClick = () => {
        const kurtosisPackage = kurtosisPackages[chosenPackage]
        navigate("/catalog/form", {state: {kurtosisPackage}})
    }

    const handleSearchEvent = (e) => {
        const value = e.target.value
        if (value === "") {
            setKurtosisPackages(defaultPackages)
        }
        const filteredPackages = defaultPackages.filter(pack => {
                if ("name" in pack) {
                    return pack.name.includes(value)
                }
                return false;
            }
        )
        
        setKurtosisPackages(filteredPackages)
    }

    return (
        <div className='w-screen'>
            <Grid
                templateAreas={`"option"
                                "search"
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
                <GridItem area={'search'} m="10px">
                    <InputGroup>
                        <InputLeftElement pointerEvents='none' >
                            <SearchIcon color='gray.300'/>
                        </InputLeftElement>
                        <Input placeholder='Package Name' color='gray.300' onChange={handleSearchEvent}/>
                    </InputGroup>
                </GridItem>
                <GridItem area={'main'} h="100%" overflowY={"scroll"}> 
                    <List spacing={1} padding="10px" h="100%">
                        {
                            kurtosisPackages.map( (kurtosisPackage, index) => {
                                const bgcolor = (index === chosenPackage) ? '#24BA27' : 'gray.300'
                                if ("name" in kurtosisPackage) {
                                    return (
                                        <ListItem bg={bgcolor} key={index} onClick={() => selectPackage(index)}>
                                            <Center h="70px" w="100%">
                                                <Text fontSize={"2xl"} color='blue.800' fontWeight={"bold"}> {kurtosisPackage.name} </Text>
                                            </Center>
                                        </ListItem> 
                                    )
                                }
                            })
                        }
                    </List>
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Button bg='#24BA27' w="100%" isDisabled={chosenPackage === null} onClick={handleConfigureButtonClick} >Configure >> </Button>
                </GridItem>
            </Grid>
        </div>
    );
  };
  export default PackageCatalog;






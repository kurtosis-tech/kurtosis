import {useEffect, useState} from "react";
import {
    Box,
    Button,
    Center,
    Grid,
    GridItem,
    Input,
    InputGroup,
    InputLeftElement,
    List,
    ListItem,
    Stack,
    Text,
    Tooltip
} from '@chakra-ui/react'

import {SearchIcon} from '@chakra-ui/icons'
import PackageCatalogOption from "./PackageCatalogOption";
import PackageCatalogMainComponent from "./PackageCatalogMainComponent";
import {useNavigate} from "react-router";
import {getOwnerNameFromUrl} from "../api/packageCatalog";
import Moment from "react-moment";

const PackageCatalog = ({kurtosisPackages: defaultPackages}) => {
    const navigate = useNavigate()
    const [kurtosisPackages, setKurtosisPackages] = useState([])
    const [chosenPackage, setChosenPackage] = useState({})

    useEffect(() => {
        setKurtosisPackages(defaultPackages)
    }, [defaultPackages.length])

    const selectPackage = (selectedPackage) => {
        if (selectedPackage["name"] === chosenPackage["name"]) {
            setChosenPackage({})
        } else {
            setChosenPackage(selectedPackage)
        }
    }

    const handleConfigureButtonClick = () => {
        navigate("/catalog/create", {state: {kurtosisPackage: chosenPackage}})
    }

    const handleSearchEvent = (e) => {
        const value = e.target.value
        if (value === "") {
            setKurtosisPackages(defaultPackages)
        }
        const filteredPackages = defaultPackages.filter(pack => {
                if ("name" in pack) {
                    // lowercase everything so that it works for both cases
                    const trimmedValue = value.trim()
                    return pack.name.toLowerCase().includes(trimmedValue)
                }
                return false;
            }
        )
        setKurtosisPackages(filteredPackages)
    }

    const getOwner = (packageName) => {
        const {owner} = getOwnerNameFromUrl(packageName)
        return owner
    }

    const getName = (packageName) => {
        const {name} = getOwnerNameFromUrl(packageName)
        return name
    }

    const renderKurtosisPackages = () => (
        kurtosisPackages.map((kurtosisPackage, index) => {
            const bgcolor = (kurtosisPackage.name === chosenPackage.name) ? '#24BA27' : 'gray.300'
            if ("name" in kurtosisPackage) {
                const date = new Date(kurtosisPackage.parsingTime)
                return (
                    <ListItem bg={bgcolor} key={index} onClick={() => selectPackage(kurtosisPackage)}>
                        <Center h="60px" w="100%">
                            <Stack>
                                <Tooltip label={
                                    <Box>
                                        Indexed <Moment fromNow={true} date={date.toJSON()}/>
                                        <br />
                                        (<Moment fromNow={true} format="YYYY-MM-DD HH:mm:ss" date={date.toJSON()}/>)
                                    </Box>
                                }>
                                    <Text
                                        fontSize={"sm"}
                                        color='blue.800'
                                        fontWeight={"bold"}
                                    >
                                        {getOwner(kurtosisPackage.name)} / {getName(kurtosisPackage.name)}
                                    </Text>
                                </Tooltip>
                            </Stack>
                        </Center>
                    </ListItem>
                )
            }
        })
    )

    return (
        <div className='w-screen'>
            <Grid
                templateAreas={
                    `
                    "option"
                    "search"
                    "main"
                    "configure"
                    "load"
                    `
                }
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
                <GridItem area={'search'} m="10px">
                    <InputGroup>
                        <InputLeftElement pointerEvents='none'>
                            <SearchIcon color='gray.300'/>
                        </InputLeftElement>
                        <Input placeholder='Package Name' color='gray.300' onChange={handleSearchEvent}/>
                    </InputGroup>
                </GridItem>
                <GridItem area={'main'} h="100%" overflowY={"scroll"}>
                    {
                        "name" in chosenPackage ?
                            <PackageCatalogMainComponent
                                renderKurtosisPackages={renderKurtosisPackages}
                                selectedKurtosisPackage={chosenPackage}
                            />
                            : <List spacing={1} padding="10px" h="100%">
                                {renderKurtosisPackages()}
                            </List>
                    }
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Button bg='#24BA27' w="100%" isDisabled={Object.keys(chosenPackage).length === 0}
                            onClick={handleConfigureButtonClick}>Configure >> </Button>
                </GridItem>
            </Grid>
        </div>
    );
};
export default PackageCatalog;






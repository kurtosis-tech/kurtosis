import { 
    Grid, 
    GridItem, 
    Center 
} from '@chakra-ui/react'

import { useNavigate, useLocation} from 'react-router';

const PackageCatalogOption = ({catalog}) => {
    const navigate = useNavigate();
    const handleCatalogForm = () => {
        if (!catalog) {
            navigate("/catalog")
        }
    }

    return (
        <Grid
            templateAreas={`"catalog manual"`}
            gridTemplateRows={'1fr'}
            gridTemplateColumns={'1fr 1fr'}
            h='100%'
            w='100%'
            color='blackAlpha.700'
            fontWeight='bold'
            gap={2}
        >
            <GridItem area={'catalog'}>
                <Center border={catalog ? "2px": null} h="100%" w="100%" color='white' bg="#24BA27" p="2" onClick={handleCatalogForm}> 
                    Catalog
                </Center>
            </GridItem>
            <GridItem area={'manual'}>
                <Center h="100%" w="100%" color='white' bg="#24BA27" p="2" onClick={ () => navigate("/enclave/create")}>
                    Manual
                </Center>
            </GridItem>
        </Grid>
    )
}

export default PackageCatalogOption;

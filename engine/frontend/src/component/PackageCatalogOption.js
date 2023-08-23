import { 
    Grid, 
    GridItem, 
    Center 
} from '@chakra-ui/react'

const PackageCatalogOption = () => {
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
                <Center h="100%" w="100%" color='white' bg="green.600" p="2">
                    Catalog
                </Center>
            </GridItem>
            <GridItem area={'manual'}>
                <Center h="100%" w="100%" color='white' bg="green.600" p="2">
                    Manual
                </Center>
            </GridItem>
        </Grid>
    )
}

export default PackageCatalogOption;

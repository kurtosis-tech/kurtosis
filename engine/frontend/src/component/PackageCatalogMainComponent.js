import {Grid, GridItem, List,} from '@chakra-ui/react'

import PackageCatalogDescription from "./PackageCatalogDescription";

const PackageCatalogMainComponent = ({renderKurtosisPackages, selectedKurtosisPackage}) => {
    return (
        <Grid
            templateAreas={`"enclaves description"`}
            gridTemplateRows={'1fr'}
            gridTemplateColumns={'1fr 1fr'}
            h='100%'
            w='100%'
            color='blackAlpha.700'
            fontWeight='bold'
            gap={2}
        >
            <GridItem area={'enclaves'} h="100%" overflowY={"scroll"}>
                <List spacing={1} padding="10px" h="100%">
                    {renderKurtosisPackages()}
                </List>
            </GridItem>
            <GridItem area={'description'} h="100%" overflowY={"scroll"} className="bg-white">
                {
                    selectedKurtosisPackage.description ? (
                        <PackageCatalogDescription
                            content={selectedKurtosisPackage.description}
                            lastIndexedTimestamp={selectedKurtosisPackage.parsingTime}
                            sha={selectedKurtosisPackage.version}
                        />
                    ) : <PackageCatalogDescription content={`No Description found for: ${selectedKurtosisPackage.name}`}/>
                }
            </GridItem>
        </Grid>
    )
}


export default PackageCatalogMainComponent

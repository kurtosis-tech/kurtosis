import { Center, SimpleGrid, Heading, Text, Box } from '@chakra-ui/react';
import React from 'react';
import { Link } from "react-router-dom";

const TitleBar = () => {
    return (
        <SimpleGrid columns={3} spacing={1} paddingBottom={5}>
            <Link to="/">
                <Box 
                    height={"40px"} 
                    justify={'flex-start'}> 
                        <svg className="m-2 w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 20">
                            <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 8v10a1 1 0 0 0 1 1h4v-5a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v5h4a1 1 0 0 0 1-1V8M1 10l9-9 9 9"/>
                        </svg>
                </Box>
            </Link>
            <Heading color={"#24BA27"}>
                <Center>
                    <Text> Kurtosis Enclave Manager </Text>
                </Center>
            </Heading>  
        </SimpleGrid>
    );
};

export default TitleBar;

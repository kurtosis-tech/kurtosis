import './App.css';
import Home from './component/Home';
import { ChakraProvider } from '@chakra-ui/react'

const App = () => {
  console.log("Enclave Manager version: 2023-08-28-4")
  return (
    <ChakraProvider>
      <div className="h-screen w-screen">
          <Home />
      </div> 
    </ChakraProvider>  
  )
}

export default App;

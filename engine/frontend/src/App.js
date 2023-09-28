import './App.css';
import Home from './component/Home';
import { ChakraProvider } from '@chakra-ui/react'

const App = () => {
  console.log("Enclave Manager version: 2023-09-27-01")
  return (
    <ChakraProvider>
      <div className="h-screen w-screen bg-[#171923]">
          <Home />
      </div> 
    </ChakraProvider>  
  )
}

export default App;

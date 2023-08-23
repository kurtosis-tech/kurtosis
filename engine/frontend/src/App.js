import './App.css';
import Home from './component/Home';
import { BrowserRouter as Router } from 'react-router-dom';
import { ChakraProvider } from '@chakra-ui/react'

const App = () => {
  return (
    <ChakraProvider>
      <div className="h-screen w-screen">
        <Router>
          <Home />
        </Router>
      </div> 
    </ChakraProvider>  
  )
}

export default App;

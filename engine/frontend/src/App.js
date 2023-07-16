import './App.css';
import Home from './components/Home';
import Temp from './component/Temp';
import Services from "./components/Services"
import ServiceInfo from "./components/ServiceInfo"
import CreateEnclave from './components/CreateEnclave';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

const App = () => {
  return (
      <div className="h-screen w-screen">
        <Router>
          <Temp />
        </Router>

        {/* <Routes>
            <Route exact path="/" element={} />
            <Route path="/create" element={<CreateEnclave />} />
            <Route path="/enclaves/:name/services/:uuid" element={<ServiceInfo/>} />
            <Route path="/enclaves/:name" element={<Services/>} />
        </Routes> */}
      </div>
    
  )
}

export default App;

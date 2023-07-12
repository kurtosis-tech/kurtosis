import './App.css';
import Home from './components/Home';
import Services from "./components/Services"
import ServiceInfo from "./components/ServiceInfo"
import CreateEnclave from './components/CreateEnclave';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

const App = () => {
  return (
    <Router>
      <div className="h-screen w-screen">
        <Routes>
            <Route exact path="/" element={<Home />} />
            <Route path="/create" element={<CreateEnclave />} />
            <Route path="/enclaves/:name/services/:uuid" element={<ServiceInfo/>} />
            <Route path="/enclaves/:name" element={<Services/>} />
        </Routes>
      </div>
    </Router>
    
  )
}

export default App;

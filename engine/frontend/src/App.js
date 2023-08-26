import './App.css';
import Home from './component/Home';
import {BrowserRouter as Router} from 'react-router-dom';

const App = () => {
    console.log("Enclave Manager version: 2023-08-26-1")
    return (
        <div className="h-screen w-screen">
            <Router>
                <Home/>
            </Router>
        </div>
    )
}

export default App;

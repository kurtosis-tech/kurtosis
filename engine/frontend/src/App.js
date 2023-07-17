import './App.css';
import Temp from './component/Temp';
import { BrowserRouter as Router } from 'react-router-dom';

const App = () => {
  return (
      <div className="h-screen w-screen">
        <Router>
          <Temp />
        </Router>
      </div>   
  )
}

export default App;

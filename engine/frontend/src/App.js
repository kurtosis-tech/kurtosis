import './App.css';
import Home from './component/Home';
import { BrowserRouter as Router } from 'react-router-dom';

const App = () => {
  return (
      <div className="h-screen w-screen">
        <Router>
          <Home />
        </Router>
      </div>   
  )
}

export default App;

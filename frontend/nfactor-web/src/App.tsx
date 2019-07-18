import React from 'react';
import { BrowserRouter as Router, Route} from "react-router-dom";

import './App.css';
import {PublicLayout} from './layouts/PublicLayout';
import {LoginPage} from './pages/login/LoginPage';


const App: React.FC = () => {
  
  return (
    <div className="App">
        <Router>
        <div className="router">
          <Route exact path="/" component={withLayout(<LoginPage />)} />
        </div>
      </Router>
    </div>
  );
}

export default App;

const withLayout = (page: React.ReactNode) => {
    return(
      () => <PublicLayout 
        header={"Header"}
        content={page} 
        footer={"N-Factor Auth, 2019"}   
      />
    );
  }
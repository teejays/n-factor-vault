import React from 'react';
import logo from './logo.svg';
import './App.css';

import {Typography, Button, Row} from 'antd';

import {PublicLayout} from './layouts/PublicLayout';
import {LoginForm} from './components/LoginForm';

const App: React.FC = () => {
  return (
    <div className="App">
      <PublicLayout 
      header={"Header"}
      content={<LoginFormComponent />} 
      footer={"N-Factor Auth, 2019"}   
    />
    </div>
  );
}

const LoginFormComponent = () => {
  const { Title } = Typography;
  return (
    <Row type="flex" justify="center" align="middle" style={{textAlign: 'center', minHeight: 600}}>
      <LoginForm 
        handleLogin={()=>{}}
        maxWidth={300}
      />
    </Row>
  )
}

export default App;

const DevWelcomeComponent = () => {
  return (
     <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.tsx</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
        <div>
          <Button type="primary">Button</Button>
        </div>
    </header>
  )
};

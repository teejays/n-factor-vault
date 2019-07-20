import React from 'react';
import {BrowserRouter as Router, Route} from 'react-router-dom';

import './App.css';
import {PublicLayout} from './layouts/PublicLayout';
import {LoginPage} from './pages/login/LoginPage';

const App: React.FC = () => {
  return (
    // This is the entry point for the app.
    // Let's only put the routes here, and then put the pages
    // in the pages directory, and components shareable among pages
    // in the components directory
    <div className="App">
      <Router>
        <div className="router">
          <Route exact path="/" component={withLayout(<LoginPage />)} />
        </div>
      </Router>
    </div>
  );
};

export default App;

// withLayout is a higher order component (HOC), it basically
// wraps around a component (page) and put's it in a 'layout',
// with  a standard header and footer.
const withLayout = (page: React.ReactNode) => {
  return () => (
    <PublicLayout
      header={'Header'}
      content={page}
      footer={'N-Factor Auth, 2019'}
    />
  );
};

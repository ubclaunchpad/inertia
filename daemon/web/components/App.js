import React from 'react';
import { Redirect, Router, Route } from 'react-router-dom';
import PropTypes from 'prop-types';

import InertiaClient from '../client';
import Login from './Login';
import Home from './Home';
import { isAuthenticated } from '../common/AuthService';

const AuthRoute = ({ component: Component, ...rest }) => (
  <Route {...rest} render={(props) => (
    isAuthenticated()
      ? <Component {...props}/>
      : <Redirect to="/login"/>
  )}/>
);

export default class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <Router>
        <Route path="/login" component={Login}/>
        <AuthRoute path="/home" component={Home}/>
      </Router>
    );
  }
}

App.propTypes = {
  client: PropTypes.instanceOf(InertiaClient)
};

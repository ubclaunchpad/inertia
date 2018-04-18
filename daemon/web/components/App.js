import React from 'react';
import { Redirect, HashRouter, Route } from 'react-router-dom';
import PropTypes from 'prop-types';

import InertiaClient from '../client';
import Login from './Login';
import Home from './Home';

const isAuthenticated = () => {
  // TODO: authentication check prior to route change
  return true;
};

const AuthRoute = ({ component: Component, props, ...rest }) => (
  <Route {...rest} render={(routeProps) => (
    isAuthenticated()
      ? <Component {...Object.assign({}, routeProps, props)}/>
      : <Redirect to="/login"/>
  )}/>
);

// render a route component with props
const PropsRoute = ({ component: Component, props, ...rest }) => (
  <Route {...rest} render={(routeProps) => (
      <Component {...Object.assign({}, routeProps, props)}/>
  )}/>
);

export default class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <HashRouter>
        <div style={styles.container}>
          <Route exact path="/" component={() => <Redirect to="/login" />}/>
          <PropsRoute path="/login" component={Login} props={this.props}/>
          <AuthRoute path="/home" component={Home} props={this.props}/>
        </div>
      </HashRouter>
    );
  }
}

App.propTypes = {
  client: PropTypes.instanceOf(InertiaClient)
};

const styles = {
  container: {
    display: 'flex',
    height: '100%',
    width: '100%'
  }
};

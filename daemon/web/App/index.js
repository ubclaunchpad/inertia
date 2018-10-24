import React from 'react';
import { HashRouter, Redirect, Route } from 'react-router-dom';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';

import Main from './Main';
import Login from '../pages/Login';

const styles = {
  container: {
    display: 'flex',
    height: '100%',
    width: '100%',
  },
};

// render a route component that requires authentication
const AuthRoute = ({ authenticated, component: Component, props, ...rest }) => (
  <Route
    {...rest}
    render={routeProps => (
      authenticated
        ? <Component {...Object.assign({}, routeProps, props)} />
        : <Redirect to="/login" />
    )} />
);
AuthRoute.propTypes = {
  authenticated: PropTypes.bool,
  component: PropTypes.any,
  props: PropTypes.shape(),
};


// render a route component with props
const PropsRoute = ({ component: Component, props, ...rest }) => (
  <Route
    {...rest}
    render={routeProps => (<Component {...Object.assign({}, routeProps, props)} />
    )} />
);
PropsRoute.propTypes = {
  component: PropTypes.any,
  props: PropTypes.shape(),
};

const AppWrapper = ({ authenticated }) => (
  <HashRouter>
    <div style={styles.container}>
      <Route
        exact
        path="/"
        component={() => (authenticated
          ? <Redirect to="/app" />
          : <Redirect to="/login" />)} />
      <PropsRoute
        path="/login"
        component={authenticated
          ? () => <Redirect to="/app" />
          : Login} />
      <AuthRoute
        path="/app"
        component={Main}
        authenticated={authenticated} />
    </div>
  </HashRouter>
);
AppWrapper.propTypes = { authenticated: PropTypes.bool };

const mapStateToProps = ({ Auth: { authenticated } }) => ({ authenticated });

const App = connect(mapStateToProps)(AppWrapper);

export default App;

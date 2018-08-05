import React from 'react';
import { HashRouter, Redirect, Route } from 'react-router-dom';
import PropTypes from 'prop-types';
import createHistory from 'history/createBrowserHistory';

import Login from '../login/Login';
import Main from '../main/Main';

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

export default class App extends React.Component {
  static async isAuthenticated() {
    // TODO: disable route guards
    // const response = await InertiaAPI.validate();
    // return (response.status === 200);
    return true;
  }

  constructor(props) {
    super(props);

    this.state = {
      loading: true,
      authenticated: false,
    };

    this.isAuthenticated = App.isAuthenticated.bind(this);

    this.isAuthenticated()
      .then((authenticated) => {
        this.setState({
          loading: false,
          authenticated,
        });
      });

    const history = createHistory();
    history.listen(() => {
      this.setState({ loading: true });
      this.isAuthenticated()
        .then((authenticated) => {
          this.setState({
            loading: false,
            authenticated,
          });
        });
    });
  }

  render() {
    if (this.state.loading) {
      return (
        <p align="center">
          Loading...
        </p>
      );
    }

    return (
      <HashRouter>
        <div style={styles.container}>
          <Route
            exact
            path="/"
            component={() => <Redirect to="/login" />} />
          <PropsRoute
            path="/login"
            component={Login}
            props={this.props} />
          <AuthRoute
            path="/home"
            authenticated={this.state.authenticated}
            component={Main}
            props={this.props} />
        </div>
      </HashRouter>
    );
  }
}

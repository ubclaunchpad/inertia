import React from 'react';
import { Redirect, HashRouter, Route } from 'react-router-dom';
import PropTypes from 'prop-types';

import InertiaClient from '../client';
import Login from './Login';
import Home from './Home';

const AuthRoute = ({ authenticated, component: Component, props, ...rest }) => (
  <Route {...rest} render={(routeProps) => (
    authenticated
      ? <Component {...Object.assign({}, routeProps, props)} />
      : <Redirect to="/login" />
  )} />
);

// render a route component with props
const PropsRoute = ({ component: Component, props, ...rest }) => (
  <Route {...rest} render={(routeProps) => (
    <Component {...Object.assign({}, routeProps, props)} />
  )} />
);

export default class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,
      authenticated: false,
    };
    this.isAuthenticated = this.isAuthenticated.bind(this)
  }

  async isAuthenticated() {
    const params = {
      headers: { 'Accept': 'application/json' },
    };
    const response = await this.props.client._get("/validate", params);
    return (response.status === 200);
  };

  componentDidMount() {
    this.isAuthenticated().then((authenticated) => {
      this.setState({
        loading: false,
        authenticated,
      });
    });
  }

  render() {
    const { component: Component, ...rest } = this.props;
    if (this.state.loading) {
      return <div>Loading...</div>;
    } else {
      return (
        <HashRouter>
          <div style={styles.container}>
            <Route exact path="/" component={() => <Redirect to="/login" />} />
            <PropsRoute path="/login" component={Login} props={this.props} />
            <AuthRoute path="/home"
              authenticated={this.state.authenticated}
              component={Home} props={this.props} />
          </div>
        </HashRouter>
      )
    }
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

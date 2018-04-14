import React from 'react';
import PropTypes from 'prop-types';
import InertiaClient from '../client';
import App from './App';

export default class Login extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: "",
      password: "",
      loginAlert: ""
    };
    this.handleLoginSubmit = this.handleLoginSubmit.bind(this);
    this.handleGetLogs = this.handleGetLogs.bind(this);
    this.handleUsernameBlur = this.handleUsernameBlur.bind(this);
    this.handlePasswordBlur = this.handlePasswordBlur.bind(this);
  }

  async handleLoginSubmit() {
    const endpoint = '/web/login';
    const params = {
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: 'bear',
        password: 'tree',
        // username: this.state.username,
        // password: this.state.password,
        email: "",
        admin: true
      })
    };

    const response = await this.props.client._post(endpoint, params);
    console.log("Login response is: ", response);
    // this.setState({loginAlert: response.body});
  }

  async handleGetLogs() {
    const endpoint = '/logs';
    const params = {
      headers: {
        'Accept': 'application/json'
      }
    };

    const response = await this.props.client._get(endpoint, params);
    console.log("Logs response is: ", response);
  }

  handleUsernameBlur(e) {
    this.setState({username: e.target.value});
  }

  handlePasswordBlur(e) {
    this.setState({password: e.target.value});
  }

  render () {
    return (
      <div>
        <p align="center">
          <img
            src="https://github.com/ubclaunchpad/inertia/blob/master/.static/inertia-with-name.png?raw=true"
            width="10%"/>
        </p>
        <p align="center">
          This is the Inertia web client!
        </p>
        <div style={styles.login}>
          <input onBlur={this.handleUsernameBlur} placeholder="Username"/>
          <input onBlur={this.handlePasswordBlur} placeholder="Password"/>
          <button onClick={this.handleLoginSubmit}>Login</button>
          <p>{this.state.loginAlert}</p>
          <button onClick={this.handleGetLogs}>Get Logs</button>
        </div>
      </div>
    )
  }
}

App.propTypes = {
  client: PropTypes.instanceOf(InertiaClient)
};

const styles = {
  container: {
    display: 'flex',
  },
  login: {
    display: 'flex',
    flexFlow: 'column',
    alignItems: 'center'
  }
};


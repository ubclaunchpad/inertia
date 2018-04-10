import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../client';

export default class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: "",
      password: "",
      loginAlert: ""
    };
    this.handleLoginSubmit = this.handleLoginSubmit.bind(this);
    this.handleUsernameBlur = this.handleUsernameBlur.bind(this);
    this.handlePasswordBlur = this.handlePasswordBlur.bind(this);
  }

  async handleLoginSubmit() {
    const endpoint = '/web/login';
    const params = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: this.state.username,
        password: this.state.password,
        email: "",
        admin: false
      })
    };

    const response = await this.props.client._post(endpoint, params).json();
    this.setState({loginAlert: response.body});
  }

  handleUsernameBlur(e) {
    this.setState({username: e.target.value});
  }

  handlePasswordBlur(e) {
    this.setState({password: e.target.value});
  }

  render() {
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
        </div>
      </div>
    );
  }
}

App.propTypes = {
  client: PropTypes.instanceOf(InertiaClient)
}

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

import React from 'react';

export default class Login extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: '',
      password: '',
      loginAlert: ''
    };
    this.handleLoginSubmit = this.handleLoginSubmit.bind(this);
    this.handleUsernameBlur = this.handleUsernameBlur.bind(this);
    this.handlePasswordBlur = this.handlePasswordBlur.bind(this);
  }

  async handleLoginSubmit() {
    const response = await this.props.client.login(
      this.state.username,
      this.state.password
    );
    if (response.status !== 200) {
      this.setState({ loginAlert: 'Username and/or password is incorrect' });
      return;
    }
    this.props.history.push('/home');
  }

  handleUsernameBlur(e) {
    this.setState({ username: e.target.value });
  }

  handlePasswordBlur(e) {
    this.setState({ password: e.target.value });
  }

  render() {
    return (
      <div style={styles.container}>
        <p align="center">
          <img
            src="https://github.com/ubclaunchpad/inertia/blob/master/.static/inertia-with-name.png?raw=true"
            width="20%" />
        </p>
        <div style={styles.login}>
          <input onBlur={this.handleUsernameBlur} placeholder="Username" />
          <input type='password'　style={{ marginBottom: '0.5rem' }} onBlur={this.handlePasswordBlur} placeholder="Password" />
          <button onClick={this.handleLoginSubmit}>Login</button>
          <p style={styles.loginAlert}>{this.state.loginAlert}</p>
        </div>
      </div>
    );
  }
}

const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    justifyContent: 'center',
    height: '100%',
    width: '100%'
  },

  login: {
    position: 'relative',
    display: 'flex',
    flexFlow: 'column',
    alignItems: 'center',
    margin: '0.5rem 0',
    marginBottom: '10rem'
  },

  loginAlert: {
    position: 'absolute',
    top: '105%'
  }
};


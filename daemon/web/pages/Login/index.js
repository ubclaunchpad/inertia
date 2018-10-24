import React from 'react';
import PropTypes from 'prop-types';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';

// import InertiaAPI from '../../common/API';
import * as loginActions from '../../actions/auth/login';

const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    justifyContent: 'center',
    height: '100%',
    width: '100%',
  },

  login: {
    position: 'relative',
    display: 'flex',
    flexFlow: 'column',
    alignItems: 'center',
    margin: '0.5rem 0',
    marginBottom: '10rem',
  },
};

class LoginWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: '',
      password: '',
    };
    this.handleLoginSubmit = this.handleLoginSubmit.bind(this);
    this.handleUsernameBlur = this.handleUsernameBlur.bind(this);
    this.handlePasswordBlur = this.handlePasswordBlur.bind(this);
  }

  async handleLoginSubmit() {
    const { loginAction } = this.props;
    const { username, password } = this.state;
    loginAction({ username, password });
  }

  handleUsernameBlur(e) {
    this.setState({ username: e.target.value });
  }

  handlePasswordBlur(e) {
    this.setState({ password: e.target.value });
  }

  render() {
    const { error = {}, authenticated, history } = this.props;
    console.log({ authenticated, error });

    if (authenticated) history.push('/app');

    return (
      <div style={styles.container}>
        <p align="center">
          <img
            alt="logo"
            src="https://github.com/ubclaunchpad/inertia/blob/master/.static/inertia-with-name.png?raw=true"
            width="20%"
          />
        </p>
        <div style={styles.login}>
          <input onBlur={this.handleUsernameBlur} placeholder="Username" />
          <input
            type="password"
            style={{ marginBottom: '0.5rem' }}
            onBlur={this.handlePasswordBlur}
            placeholder="Password"
          />
          <button type="submit" onClick={this.handleLoginSubmit}>
Login
          </button>

          <br />
          <h2>
            {error.message || ''}
          </h2>
        </div>
      </div>
    );
  }
}
LoginWrapper.propTypes = {
  history: PropTypes.object,
  loginAction: PropTypes.func,
  authenticated: PropTypes.bool.isRequired,
  error: PropTypes.any,
};

const mapStateToProps = ({ Auth: { authenticated, error } }) => ({ authenticated, error });

const mapDispatchToProps = dispatch => bindActionCreators({ ...loginActions }, dispatch);

const Login = connect(mapStateToProps, mapDispatchToProps)(LoginWrapper);

export default Login;

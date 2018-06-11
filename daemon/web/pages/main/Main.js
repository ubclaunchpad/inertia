import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Link,
  Route,
  Switch,
} from 'react-router-dom';

import InertiaAPI from '../../common/API';
import Containers from '../containers/Containers';
import Dashboard from '../dashboard/Dashboard';
import * as mainActions from '../../actions/main';

// hardcode all styles for now, until we flesh out UI
const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    height: '100%',
    width: '100%',
  },

  innerContainer: {
    display: 'flex',
    flexFlow: 'row',
    height: '100%',
    width: '100%',
  },

  headerBar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    width: '100%',
    height: '4rem',
    padding: '0 1.5rem',
    borderBottom: '1px solid #c1c1c1',
  },

  sidebar: {
    display: 'flex',
    flexFlow: 'column',
    width: '20rem',
    height: '100%',
    paddingTop: '0.5rem',
    borderRight: '1px solid #c1c1c1',
    backgroundColor: '#f0f0f0',
  },

  main: {
    height: '100%',
    width: '100%',
    overflowY: 'scroll',
  },

  button: {
    flex: 'none',
  },
};

const sidebarHeaderStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '1.5rem',
    paddingTop: '1rem',
  },

  text: {
    textDecoration: 'none',
    color: '#5f5f5f',
  },
};

const sidebarTextStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: 'flex',
    width: '100%',
    paddingLeft: '2rem',
    paddingTop: '0.5rem',
  },

  text: {
    fontSize: '80%',
    textDecoration: 'none',
    color: '#101010',
  },

  button: {
    fontSize: '80%',
    textDecoration: 'none',
    color: '#101010',
  },
};

const SidebarHeader = ({ children, onClick }) => (
  <div style={sidebarHeaderStyles.container}>
    <button onClick={onClick} style={sidebarHeaderStyles.text}>
      {children}
    </button>
  </div>
);
SidebarHeader.propTypes = {
  children: PropTypes.node,
  onClick: PropTypes.func,
};

const SidebarButton = ({ children, onClick }) => (
  <div style={sidebarTextStyles.container}>
    <button onClick={onClick} style={sidebarTextStyles.button}>
      {children}
    </button>
  </div>
);
SidebarButton.propTypes = {
  children: PropTypes.node,
  onClick: PropTypes.func,
};

const SidebarText = ({ children }) => (
  <div style={sidebarTextStyles.container}>
    <p style={sidebarTextStyles.text}>
      {children}
    </p>
  </div>
);
SidebarText.propTypes = {
  children: PropTypes.node,
};

class MainWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      remoteVersion: '',

      repoBranch: '',
      repoCommitHash: '',
      repoCommitMessage: '',
      repoBuildType: '',
      repoBuilding: false,
      containers: [],

      viewContainer: '',
    };

    this.handleLogout = this.handleLogout.bind(this);
    this.handleGetStatus = this.handleGetStatus.bind(this);

    this.handleGetStatus()
      .then(() => {})
      .catch(() => {}); // TODO: Log error
  }

  async handleLogout() {
    const response = await InertiaAPI.logout();
    if (response.status !== 200) {
      // TODO: Log Error
      return;
    }
    this.props.history.push('/login');
  }

  async handleGetStatus() {
    const response = await InertiaAPI.getRemoteStatus();
    if (response.status !== 200) return new Error('bad response: ' + response);
    const status = await response.json();
    this.setState({
      remoteVersion: status.version,
      repoBranch: status.branch,
      repoBuilding: status.build_active,
      repoBuildType: status.build_type,
      repoCommitHash: status.commit_hash,
      repoCommitMessage: status.commit_message,
      containers: status.containers,
    });
    return null;
  }

  render() {
    // Render container list
    const containers = this.state.containers.map(c => (
      <SidebarButton
        onClick={() => { this.setState({ viewContainer: c }); }}
        key={c}>
        <code>{c}</code>
      </SidebarButton>
    ));

    // Report repository status
    const buildMessage = this.state.repoBuilding
      ? <SidebarText>Build in progress</SidebarText>
      : null;
    const repoState = this.state.repoCommitHash
      ? (
        <div>
          <SidebarText>Type: <code>{this.state.repoBuildType}</code></SidebarText>
          <SidebarText>Branch: <code>{this.state.repoBranch}</code></SidebarText>
          <SidebarText>Commit: <code>{this.state.repoCommitHash.substr(1, 8)}</code>
            <br />&quot;{this.state.repoCommitMessage}&quot;
          </SidebarText>
        </div>
      )
      : null;

    return (
      <div style={styles.container}>

        <header style={styles.headerBar}>
          <p style={{
            fontWeight: 500,
            fontSize: 24,
            color: '#101010',
          }}>
            Inertia Web
          </p>

          <button
            onClick={this.handleLogout}
            style={{
              textDecoration: 'none',
              color: '#5f5f5f',
            }}>
            logout
          </button>

          <Link to={`${this.props.match.url}/dashboard`}>Click to go to Dashboard</Link>
          <Link to={`${this.props.match.url}/containers`}>Click to go to Containers</Link>
        </header>

        <div style={styles.innerContainer}>

          <div style={styles.sidebar}>
            <SidebarHeader onClick={() => { this.setState({ viewContainer: '' }); }}>
              Daemon
            </SidebarHeader>
            <SidebarText><code>{this.state.remoteVersion}</code></SidebarText>
            <SidebarHeader>Repository Status</SidebarHeader>
            {buildMessage}
            {repoState}
            <SidebarHeader>Active Containers</SidebarHeader>
            {containers}
          </div>

          <div style={styles.main}>
            <Switch>
              <Route
                exact
                path={`${this.props.match.url}/dashboard`}
                component={() => <Dashboard container={this.state.viewContainer} />}
              />
              <Route
                exact
                path={`${this.props.match.url}/containers`}
                component={() => <Containers />}
              />
            </Switch>
          </div>
        </div>
      </div>
    );
  }
}
MainWrapper.propTypes = {
  history: PropTypes.object,
  match: PropTypes.object,
};


const mapStateToProps = ({ Main }) => {
  return {
    testState: Main.testState,
  };
};

const mapDispatchToProps = dispatch => bindActionCreators({ ...mainActions }, dispatch);

const Main = connect(mapStateToProps, mapDispatchToProps)(MainWrapper);


export default Main;

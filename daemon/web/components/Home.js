import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../client';
import Dashboard from './Dashboard';

const SidebarHeader = ({ children, onClick }) => (
  <div style={sidebarHeaderStyles.container}>
    <a onClick={onClick} style={sidebarHeaderStyles.text}>
      {children}
    </a>
  </div>
);

const sidebarHeaderStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '1.5rem',
    paddingTop: '1rem'
  },

  text: {
    textDecoration: 'none',
    color: '#5f5f5f'
  }
};

const SidebarButton = ({ children, onClick }) => (
  <div style={sidebarTextStyles.container}>
    <a onClick={onClick} style={sidebarTextStyles.text}>
      {children}
    </a>
  </div>
);

const SidebarText = ({ children }) => (
  <div style={sidebarTextStyles.container}>
    <p style={sidebarTextStyles.text}>
      {children}
    </p>
  </div>
);

const sidebarTextStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: 'flex',
    width: '100%',
    paddingLeft: '2rem',
    paddingTop: '0.5rem'
  },

  text: {
    fontSize: '80%',
    textDecoration: 'none',
    color: '#101010'
  }
};

export default class Home extends React.Component {
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
      .then(() => { })
      .catch((err) => console.error(err));
  }

  async handleLogout() {
    const response = await this.props.client.logout();
    if (response.status != 200) console.error(response);
    this.props.history.push('/login');
  }

  async handleGetStatus() {
    const response = await this.props.client.getRemoteStatus();
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
  }

  render() {
    // Render container list
    const containers = this.state.containers.map((c) =>
      <SidebarButton
        onClick={() => {
          this.setState({ viewContainer: c });
        }}
        key={c} >{c}</SidebarButton>
    );

    // Report repository status
    const buildMessage = this.state.repoBuilding
      ? <SidebarText>Build in progress</SidebarText>
      : null;
    const repoState = this.state.repoCommitHash
      ? (
        <div>
          <SidebarText>Type: {this.state.repoBuildType}</SidebarText>
          <SidebarText>Branch: {this.state.repoBranch}</SidebarText>
          <SidebarText>Commit: {this.state.repoCommitHash.substr(1, 8)} "{this.state.repoCommitMessage}"</SidebarText>
        </div>
      )
      : null;

    return (
      <div style={styles.container}>

        <header style={styles.headerBar}>
          <p style={{ fontWeight: 500, fontSize: 24, color: '#101010' }}>Inertia Web</p>
          <a onClick={this.handleLogout} style={{ textDecoration: 'none', color: '#5f5f5f' }}>logout</a>
        </header>

        <div style={styles.innerContainer}>

          <div style={styles.sidebar}>
            <SidebarHeader
              onClick={() => {
                this.setState({ viewContainer: '' });
              }}>Daemon</SidebarHeader>
            <SidebarText>{this.state.remoteVersion}</SidebarText>
            <SidebarHeader>Repository Status</SidebarHeader>
            {buildMessage}
            {repoState}
            <SidebarHeader>Active Containers</SidebarHeader>
            {containers}
          </div>

          <div style={styles.main}>
            <Dashboard
              container={this.state.viewContainer}
              client={this.props.client} />
          </div>
        </div>
      </div>
    );
  }
}

Home.propTypes = {
  client: PropTypes.instanceOf(InertiaClient),
};

// hardcode all styles for now, until we flesh out UI
const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    height: '100%',
    width: '100%'
  },

  innerContainer: {
    display: 'flex',
    flexFlow: 'row',
    height: '100%',
    width: '100%'
  },

  headerBar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    width: '100%',
    height: '4rem',
    padding: '0 1.5rem',
    borderBottom: '1px solid #c1c1c1'
  },

  sidebar: {
    display: 'flex',
    flexFlow: 'column',
    width: '20rem',
    height: '100%',
    paddingTop: '0.5rem',
    borderRight: '1px solid #c1c1c1',
    backgroundColor: '#f0f0f0'
  },

  main: {
    height: '100%',
    width: '100%',
    overflowY: 'scroll'
  },

  button: {
    flex: 'none'
  },
};

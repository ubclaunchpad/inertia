import React from 'react';

const SidebarHeader = ({ children }) => (
  <div style={sidebarHeaderStyles.container}>
    <a href="/#/home" onClick={() => { return false; }} style={sidebarHeaderStyles.text}>
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
    paddingLeft: '2rem'
  },

  text: {
    textDecoration: 'none',
    color: '#5f5f5f'
  }
};

const SidebarButton = ({ children }) => (
  <div style={sidebarButtonStyles.container}>
    <a href="/#/home" onClick={() => { }} style={sidebarButtonStyles.text}>
      {children}
    </a>
  </div>
);
const sidebarButtonStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '3rem'
  },

  text: {
    textDecoration: 'none',
    color: '#101010'
  }
};


export default class Home extends React.Component {
  constructor(props) {
    super(props);
    this.handleGetLogs = this.handleGetLogs.bind(this);
    this.handleLogout = this.handleLogout.bind(this);
  }

  async handleGetLogs() {
    const endpoint = '/logs';
    const params = {
      headers: {
        'Accept': 'application/json'
      }
    };
    const response = await this.props.client._get(endpoint, params);
  }

  async handleLogout() {
    const endpoint = '/user/logout';
    const params = {
      headers: {
        'Accept': 'application/json'
      }
    };

    const response = await this.props.client._post(endpoint, params);
    this.props.history.push('/login');
  }

  render() {
    return (
      <div style={styles.container}>

        <header style={styles.headerBar}>
          <p style={{ fontWeight: 500, fontSize: 24, color: '#101010' }}>Inertia Web</p>
          <a onClick={this.handleLogout} style={{ textDecoration: 'none', color: '#5f5f5f' }}>logout</a>
        </header>

        <div style={styles.innerContainer}>

          <div style={styles.sidebar}>
            <SidebarHeader>Deployments</SidebarHeader>
            <SidebarButton>project-app</SidebarButton>
            <SidebarButton>project-db</SidebarButton>
            <SidebarButton>project-server</SidebarButton>
          </div>

          <div style={styles.main}>
            <div style={{ display: 'flex', height: '100%', alignItems: 'center', justifyContent: 'center' }}>
              <p style={styles.underConstruction}>coming soon!</p>
            </div>
          </div>

        </div>

      </div>
    );
  }
}

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
    padding: '0 2rem',
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

  underConstruction: {
    textAlign: 'center',
    fontSize: 24,
    color: '#9f9f9f'
  }
};

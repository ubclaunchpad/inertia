import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Route,
  Switch,
} from 'react-router-dom';

import * as mainActions from '../../actions/main';

import InertiaAPI from '../../common/API';

import { Containers, Dashboard, Settings } from '../../pages';

import Footer from '../../components/Footer';
import NavBar from '../../components/NavBar';

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

  main: {
    height: '100%',
    width: '100%',
    overflowY: 'scroll',
  },
};


class MainWrapper extends React.Component {
  constructor(props) {
    super(props);

    this.handleLogout = this.handleLogout.bind(this);
    this.handleGetStatus = this.handleGetStatus.bind(this);
    this.state = { status: {} };

    this.handleGetStatus()
      .then(() => {})
      .catch(() => {}); // TODO: Log error
  }

  async handleLogout() {
    const { history } = this.props;
    const response = await InertiaAPI.logout();
    if (response.status !== 200) {
      // TODO: Log Error
      return;
    }
    history.push('/login');
  }

  async handleGetStatus() {
    const response = await InertiaAPI.getRemoteStatus();
    if (response.status !== 200) {
      // TODO: Log Error
      return;
    }
    // just a stub for now
    this.setState({ status: await response.json() });
  }

  render() {
    const { match: { url } } = this.props;
    const { status } = this.state;
    console.debug(status);
    return (
      <div style={styles.container}>
        <NavBar url={url} />

        <div style={styles.innerContainer}>
          <div style={styles.main}>
            <Switch>
              <Route
                exact
                path={`${url}/dashboard`}
                component={() => <Dashboard />}
              />
              <Route
                exact
                path={`${url}/containers`}
                component={() => <Containers dateUpdated="2018-01-01 00:00" />}
              />
              <Route
                exact
                path={`${url}/settings`}
                component={() => <Settings />}
              />
            </Switch>
          </div>
        </div>

        <Footer version="v0.0.0" />
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

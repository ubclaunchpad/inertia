import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import InertiaAPI from '../../common/API';
import LogView from '../../components/LogView';
import * as dashboardActions from '../../actions/dashboard';

const styles = {
  container: {
    display: 'flex',
    height: '100%',
    alignItems: 'center',
    justifyContent: 'center',
    position: 'relative',
  },
  underConstruction: {
    textAlign: 'center',
    fontSize: 24,
    color: '#9f9f9f',
  },
};

function promiseState(p) {
  const t = {};

  return Promise.race([p, t])
    .then(v => (v === t ? 'pending' : ('fulfilled', () => 'rejected')));
}

class DashboardWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      errored: false,
      logEntries: [],
      logReader: null,
    };
    this.getLogs = this.getLogs.bind(this);
    this.getMessage = this.getMessage.bind(this);
  }

  componentDidMount() {
    this.getLogs();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.container !== this.props.container) {
      this.getLogs();
    }
  }

  async getLogs() {
    if (this.state.logReader) await this.state.logReader.cancel();
    this.setState({
      errored: false,
      logEntries: [],
      logReader: null,
    });
    try {
      let resp;
      if (!this.props.container) {
        resp = await InertiaAPI.getContainerLogs();
      } else {
        resp = await InertiaAPI.getContainerLogs(this.props.container);
      }
      if (resp.status !== 200) {
        this.setState({
          errored: true,
          logEntries: [],
        });
      }

      const reader = resp.body.getReader();
      this.setState({ logReader: reader });

      const decoder = new TextDecoder('utf-8');
      let buffer = '';
      const stream = () => promiseState(reader.closed)
        .then((s) => {
          if (s === 'pending') {
            return reader.read()
              .then((data) => {
                const chunk = decoder.decode(data.value);
                const parts = chunk.split('\n')
                  .filter(c => c);

                parts[0] = buffer + parts[0];
                buffer = '';
                if (!chunk.endsWith('\n')) {
                  buffer = parts.pop();
                }

                this.setState({
                  logEntries: this.state.logEntries.concat(parts),
                });

                return stream();
              });
          }
          return null;
        });
      stream();
    } catch (e) {
      // TODO: Log error message
    }
  }

  getMessage() {
    if (this.state.errored) {
      return <p style={styles.underConstruction}>Yikes, something went wrong</p>;
    } else if (this.state.logEntries.length === 0) {
      return <p style={styles.underConstruction}>No logs to show</p>;
    }
    return null;
  }

  render() {
    return (
      <div style={styles.container}>
        {this.getMessage()}
        <LogView logs={this.state.logEntries} />
      </div>
    );
  }
}
DashboardWrapper.propTypes = {
  container: PropTypes.string,
};

const mapStateToProps = ({ Dashboard }) => {
  return {
    testState: Dashboard.testState,
  };
};

const mapDispatchToProps = dispatch => bindActionCreators({ ...dashboardActions }, dispatch);

const Dashboard = connect(mapStateToProps, mapDispatchToProps)(DashboardWrapper);


export default Dashboard;

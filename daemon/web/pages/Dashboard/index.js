import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import * as dashboardActions from '../../actions/project';
import TerminalView from '../../components/TerminalView';
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
  TableRowExpandable,
} from '../../components/Table';
import IconHeader from '../../components/IconHeader';

const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    height: 'min-content',
    justifyContent: 'center',
    position: 'relative',
  },
  underConstruction: {
    textAlign: 'center',
    fontSize: 24,
    color: '#9f9f9f',
  },
};

class DashboardWrapper extends React.Component {
  componentWillMount() {
    const { getStatus } = this.props;
    getStatus();
  }

  render() {
    const {
      getLogs,
      logs,
      status: {
        name,
        branch,
        commit,
        message,
        buildType,
        containers,
      } = {},
    } = this.props;

    return (
      <div style={styles.container}>
        <IconHeader title={branch} type="dashboard" />

        <Table style={{ margin: '0 30px 10px 30px' }}>
          <TableHeader>
            <TableRow>
              <TableCell>
                {name}
              </TableCell>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>
Branch
              </TableCell>
              <TableCell>
                {branch}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
Commit
              </TableCell>
              <TableCell>
                {commit}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
Message
              </TableCell>
              <TableCell>
                {message}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
Build Type
              </TableCell>
              <TableCell>
                {buildType}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <Table style={{ margin: '0 30px' }}>
          <TableHeader>
            <TableRow>
              <TableCell style={{ flex: '0 0 30%' }}>
Type/Name
              </TableCell>
              <TableCell style={{ flex: '0 0 20%' }}>
Status
              </TableCell>
              <TableCell>
Last Updated
              </TableCell>
            </TableRow>
          </TableHeader>
          <TableBody>
            {containers.map(container => (
              <TableRowExpandable
                key={container.name}
                height={300}
                onClick={() => getLogs({ container: container.name })}
                panel={<TerminalView logs={logs} />}>
                <TableCell style={{ flex: '0 0 30%' }}>
Commit
                </TableCell>
                <TableCell style={{ flex: '0 0 20%' }} />
                <TableCell>
                  {commit}
                </TableCell>
              </TableRowExpandable>
            ))
            }
          </TableBody>
        </Table>
      </div>
    );
  }
}

DashboardWrapper.propTypes = {
  logs: PropTypes.array,
  status: PropTypes.shape({
    name: PropTypes.string.isRequired,
    branch: PropTypes.string.isRequired,
    commit: PropTypes.string.isRequired,
    message: PropTypes.string.isRequired,
    buildType: PropTypes.string.isRequired,
    containers: PropTypes.arrayOf(PropTypes.shape({
      name: PropTypes.string.isRequired,
      status: PropTypes.string.isRequired,
      lastUpdated: PropTypes.string.isRequired,
    })),
  }),
  getLogs: PropTypes.func,
  getStatus: PropTypes.func,
};

const mapStateToProps = ({ Dashboard }) => ({
  status: Dashboard.status,
  logs: Dashboard.logs,
});

const mapDispatchToProps = dispatch => bindActionCreators({ ...dashboardActions }, dispatch);

const Dashboard = connect(mapStateToProps, mapDispatchToProps)(DashboardWrapper);

export default Dashboard;

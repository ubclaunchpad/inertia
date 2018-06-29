import React from 'react';
import { connect } from 'react-redux';
import ShutdownButton from '../../components/ShutdownButton/ShutdownButton';

import {
  Table,
  TableCell,
  TableRow,
  TableHeader,
  TableBody,
} from '../../components/Table/Table';

class SettingsWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  render() {
    return (

      <div>
        <h1 style={{ margin: '1rem' }}>PROJECT INFORMATION</h1>
        <Table style={{ width: '90%', margin: '1rem' }}>
          <TableHeader>
            <TableRow>
              <TableCell style={{ fontWeight: '900', fontSize: '14px' }}>Project-name</TableCell>
              <TableCell />
              <TableCell />
            </TableRow>
          </TableHeader>

          <TableBody>
            <TableRow>
              <TableCell>Branch</TableCell>
              <TableCell>somebranch</TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell>Commit</TableCell>
              <TableCell>commit hash</TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell>Message</TableCell>
              <TableCell>penguin</TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell>Build Type</TableCell>
              <TableCell>docker-compose</TableCell>
              <TableCell />
            </TableRow>
          </TableBody>
        </Table>
        <ShutdownButton style={{ margin: '1rem' }} />
      </div>
    );
  }
}

SettingsWrapper.propTypes = {};

const mapStateToProps = () => { return {}; };

const mapDispatchToProps = () => { return {}; };

const Settings = connect(mapStateToProps, mapDispatchToProps)(SettingsWrapper);

export default Settings;

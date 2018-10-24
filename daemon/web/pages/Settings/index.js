import React from 'react';
import { connect } from 'react-redux';
import {
  ShutdownButton,
  PruneImageButton,
} from '../../components/buttons';
import IconHeader from '../../components/IconHeader';

import {
  Table,
  TableCell,
  TableRow,
  TableHeader,
  TableBody,
} from '../../components/Table';

class SettingsWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  render() {
    return (

      <div>
        <IconHeader type="settings" title="PROJECT INFORMATION" style={{ margin: '1rem' }} />
        <Table style={{ width: '90%', margin: '1rem' }}>
          <TableHeader>
            <TableRow>
              <TableCell style={{ fontWeight: '900', fontSize: '14px' }}>
Project-name
              </TableCell>
              <TableCell />
              <TableCell />
            </TableRow>
          </TableHeader>

          <TableBody>
            <TableRow>
              <TableCell>
Branch
              </TableCell>
              <TableCell>
somebranch
              </TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell>
Commit
              </TableCell>
              <TableCell>
commit hash
              </TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell>
Message
              </TableCell>
              <TableCell>
penguin
              </TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell>
Build Type
              </TableCell>
              <TableCell>
docker-compose
              </TableCell>
              <TableCell />
            </TableRow>
          </TableBody>
        </Table>
        <ShutdownButton style={{ margin: '1rem' }} />
        <PruneImageButton style={{ margin: '1rem' }} />
      </div>
    );
  }
}

SettingsWrapper.propTypes = {};

const mapStateToProps = () => ({});

const mapDispatchToProps = () => ({});

const Settings = connect(mapStateToProps, mapDispatchToProps)(SettingsWrapper);

export default Settings;

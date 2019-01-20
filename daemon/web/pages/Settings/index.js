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

      <div className="pad-sides-s">
        <IconHeader type="settings" title="Project Information" />
        <Table>
          <TableHeader>
            <TableRow>
              <TableCell>
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
        <div className="pad-top-s">
          <ShutdownButton />
        </div>
        <div className="pad-top-s">
          <PruneImageButton />
        </div>
      </div>
    );
  }
}

SettingsWrapper.propTypes = {};

const mapStateToProps = () => ({});

const mapDispatchToProps = () => ({});

const Settings = connect(mapStateToProps, mapDispatchToProps)(SettingsWrapper);

export default Settings;

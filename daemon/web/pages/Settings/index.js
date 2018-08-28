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
        <Table style={{ margin: '0 30px 10px 30px' }}>
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
        <ShutdownButton style={{ margin: '30px' }} />
        <PruneImageButton style={{ margin: '30px' }} />
      </div>
    );
  }
}

SettingsWrapper.propTypes = {};

const mapStateToProps = () => { return {}; };

const mapDispatchToProps = () => { return {}; };

const Settings = connect(mapStateToProps, mapDispatchToProps)(SettingsWrapper);

export default Settings;

import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';

import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from '../../components/Table';
import IconHeader from '../../components/IconHeader';
import {
  AddUserButton
} from '../../components/buttons';

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

class TeamWrapper extends React.Component {
  // componentWillMount() {
  //   const { handleGetProjectDetails, handleGetContainers } = this.props;
  //   handleGetProjectDetails();
  //   handleGetContainers();
  // }

  render() {
    const {
      teamMemebers,
    } = this.props;

    return (
      <div style={styles.container}>
        <IconHeader title="MANGE YOUR TEAM" type="team" />

        <Table style={{ margin: '0 30px 10px 30px' }}>
          <TableHeader>
            <TableRow>
              <TableCell>
Name
              </TableCell>
              <TableCell>
Role
              </TableCell>
            </TableRow>
          </TableHeader>
          <TableBody>
            asd
          </TableBody>
          {teamMemebers.map(teamMemeber => (
              <TableRow
                key={teamMemeber.name}
                height={300}>
                <TableCell style={{ flex: '0 0 30%' }}>
Commit
                </TableCell>
              </TableRow>
            ))
            }
        </Table>
        <IconHeader title="ADD USERS" type="addUser" />
        <Table style={{ margin: '0 30px' }}>
          <TableHeader>
            <TableRow>
              <TableCell style={{ flex: '0 0 30%' }}>
Username
              </TableCell>
              <input type="text" name="Username" placeholder="Enter Username"></input>
            </TableRow>
            <TableRow>
              <TableCell style={{ flex: '0 0 30%' }}>
Password
              </TableCell>
              <input type="text" name="Password" placeholder="Enter Password"></input>
            </TableRow>
            <TableRow>
              <TableCell style={{ flex: '0 0 30%' }}>
Confirm Passwordls
              </TableCell>
              <input type="text" name="ConfirmPassword" placeholder="Enter Password"></input>
            </TableRow>
          </TableHeader>
        </Table>
        <AddUserButton style={{ margin: '1rem'}} />
      </div>
    );
  }
}

TeamWrapper.propTypes = {
  teamMemebers: PropTypes.arrayOf(PropTypes.shape({
    name: PropTypes.string.isRequired,
  })),
};

const mapStateToProps = () => { return {}; };

const mapDispatchToProps = () => { return {}; };

const Team = connect(mapStateToProps, mapDispatchToProps)(TeamWrapper);

export default Team;


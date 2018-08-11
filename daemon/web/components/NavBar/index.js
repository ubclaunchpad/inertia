import React from 'react';
import PropTypes from 'prop-types';
import {
  Link,
} from 'react-router-dom';

import InertiaLogo from '../../assets/logo/inertia-horizontal-black.png';
import DashboardIcon from '../../assets/icons/dashboard-icon.png';
import TeamIcon from '../../assets/icons/team-icon.png';
import SettingIcon from '../../assets/icons/settings-icon.png';

const NavBar = ({ url }) => (
  <div>
    <img src={InertiaLogo} alt="inertia-icon" height="14" />
    <button
      type="submit"
      style={{
      textDecoration: 'none',
      color: '#5f5f5f',
    }}>
    logout
    </button>
    <Link to={`${url}/dashboard`}><img src={DashboardIcon} alt="dashboard-icon" height="12" /> Dashboard</Link>
    <Link to={`${url}/team`}><img src={TeamIcon} alt="team-icon" height="12" /> Team</Link>
    <Link to={`${url}/settings`}><img src={SettingIcon} alt="setting-icon" height="12" /> Settings</Link>
  </div>
);

NavBar.propTypes = {
  url: PropTypes.string,
};

export default NavBar;

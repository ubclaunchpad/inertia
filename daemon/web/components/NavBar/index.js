import React from 'react';
import PropTypes from 'prop-types';
import {
  Link,
} from 'react-router-dom';

import InertiaLogo from '../../assets/logo/inertia-horizontal-black.png';

import './index.sass';

const NavBar = ({ url }) => (
  <nav className="flex fill-width ai-center bg-white shadow">
    <img className="margin-sides-m " src={InertiaLogo} alt="inertia-icon" height="38" />
    <div className="flex jc-end fill-width">
      <Link to={`${url}/dashboard`} className="flex ai-center pad-sides-xs hover-color-lightred-all">
        <i className="fas fa-th-large" />
        <h1>
Dashboard
        </h1>
      </Link>
      <Link to={`${url}/team`} className="flex ai-center pad-sides-xs hover-color-lightred-all">
        <i className="fas fa-users" />
        <h1>
Team
        </h1>
      </Link>
      <Link to={`${url}/settings`} className="flex ai-center pad-sides-xs hover-color-lightred-all">
        <i className="fas fa-cog" />
        <h1>
Settings
        </h1>
      </Link>

      <button
        type="submit"
        style={{
          textDecoration: 'none',
          color: '#5f5f5f',
        }}>
    logout
      </button>
    </div>
  </nav>
);

NavBar.propTypes = {
  url: PropTypes.string,
};

export default NavBar;

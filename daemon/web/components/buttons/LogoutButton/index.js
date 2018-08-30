import React from 'react';
import PropTypes from 'prop-types';

import '../index.sass';

const LogoutButton = ({ style }) => (
  <div>
    <button className="button" type="button" style={style}>
Logout
    </button>
  </div>
);

LogoutButton.propTypes = {
  style: PropTypes.object,
};

export default LogoutButton;

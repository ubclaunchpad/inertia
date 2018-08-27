import React from 'react';
import PropTypes from 'prop-types';

import './index.sass';

const icons = {
  dashboard: <i className="fas fa-th-large" />,
  containers: <i className="fas fa-database fa-lg" />,
  team: <i className="fas fa-users" />,
  settings: <i className="fas fa-cog" />,
};

const IconHeader = ({ title, type, style }) => (
  <div className="iconheader" style={style}>
    {icons[type]}
    <h1 className="header">
      {title}
    </h1>
  </div>
);

IconHeader.propTypes = {
  title: PropTypes.string.isRequired,
  type: PropTypes.string.isRequired,
  style: PropTypes.object,
};

export default IconHeader;

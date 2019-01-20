import React from 'react';
import PropTypes from 'prop-types';

import './index.sass';

const icons = {
  dashboard: <i className="fas fa-th-large color-grey" />,
  containers: <i className="fas fa-database color-grey" />,
  team: <i className="fas fa-users color-grey" />,
  settings: <i className="fas fa-cog color-grey" />,
};

const IconHeader = ({ title, type, style }) => (
  <div className="iconheader color-white pad-top-m pad-bottom-xs" style={style}>
    {icons[type]}
    <h1 className="header color-grey pad-left-xxs">
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

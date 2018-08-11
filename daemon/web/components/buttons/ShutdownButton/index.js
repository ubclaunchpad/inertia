import React from 'react';
import PropTypes from 'prop-types';

const ShutdownButton = ({ style }) => (
  <div>
    <button className="button" type="button" style={style}>
Shut Down Containers
    </button>
  </div>
);

ShutdownButton.propTypes = {
  style: PropTypes.object,
};

export default ShutdownButton;

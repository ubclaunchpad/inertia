import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const ShutdownButton = ({ style }) => (
  <div>
    <button className="ShutdownButton" type="button" style={style}>
Shut Down Containers
    </button>
  </div>
);

ShutdownButton.propTypes = {
  style: PropTypes.object,
};

export default ShutdownButton;

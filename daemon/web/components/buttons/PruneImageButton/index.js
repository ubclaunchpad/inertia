import React from 'react';
import PropTypes from 'prop-types';

const PruneImageButton = ({ style }) => (
  <div>
    <button className="button" type="button" style={style}>
Prune Docker Image
    </button>
  </div>
);

PruneImageButton.propTypes = {
  style: PropTypes.object,
};

export default PruneImageButton;

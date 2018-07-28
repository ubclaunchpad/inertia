import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const PruneImageButton = ({ style }) =>
  (
    <div>
      <button className="PruneImageButton" type="button" style={style}>Prune Docker Image</button>
    </div>
  );

PruneImageButton.propTypes = {
  style: PropTypes.object,
};

export default PruneImageButton;

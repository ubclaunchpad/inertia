import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const Status = ({ title }) => (
  <div className="badge">
    <div className="badge badge-pill badge-primary">{title}</div>
  </div>
);

Status.propTypes = {
  title: PropTypes.string,
};

export default Status;

import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const Status = ({ title }) => (
  <div className="status">
    <span class="badge badge-pill badge-primary">{title}</span>
  </div>
);

Status.propTypes = {
  title: PropTypes.string,
};

export default Status;

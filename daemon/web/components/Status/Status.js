import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const Status = ({ title, status }) => (
  <div className="badge">
    <h3 className="title">{title}</h3>
    <div className="badge badge-pill-active">{status}</div>
  </div>
);

Status.propTypes = {
  status: PropTypes.string,
  title: PropTypes.string,
};

export default Status;

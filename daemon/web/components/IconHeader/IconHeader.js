import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const icons = {
    dashboard: <i className="fas fa-th-large"></i>,
    containers: <i className="fas fa-database fa-lg"></i>,
    team: <i className="fas fa-users"></i>,
    settings: <i className="fas fa-cog"></i>
};

const IconHeader = ({ title, type }) => (
    <div className="iconheader">
        {icons[type]}
        <h1 className="header">{title}</h1>
    </div>
);

IconHeader.propTypes = {
    title: PropTypes.string,
};

export default IconHeader;
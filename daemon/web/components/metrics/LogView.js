import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../../client';

export default class LogView extends React.Component {
    constructor(props) {
        super(props);

        this.getEntries = this.getEntries.bind(this);
    }

    getEntries() {
        let logText = '';
        for (let i = this.props.logs.length; i > 0; i--) {
            logText = logText + this.props.logs[i] + '\n';
        }
        return (
            <p>{logText}</p>
        );
    }

    render() {
        const resultList = this.getEntries();
        return (
            <div>
                {resultList}
            </div>
        );
    }
}

LogView.propTypes = {
    logs: PropTypes.array,
};

import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../../client';

export default class LogView extends React.Component {
    constructor(props) {
        super(props);

        this.getEntries = this.getEntries.bind(this);
        this.scrollToBottom = this.scrollToBottom.bind(this);
    }

    scrollToBottom() {
        this.messagesEnd.scrollIntoView({ behavior: 'smooth' });
    }

    componentDidMount() {
        this.scrollToBottom();
    }

    componentDidUpdate() {
        this.scrollToBottom();
    }

    getEntries() {
        let i = 0;
        return this.props.logs.map((l) => {
            i++;
            return (<code key={i} >{l}<br /></code>);
        });
    }

    render() {
        const resultList = this.getEntries();
        return (
            <div align='left'>
                <div>
                    {resultList}
                </div>
                <div style={{ float: 'left', clear: 'both' }}
                    ref={(el) => { this.messagesEnd = el; }}>
                </div>
            </div>
        );
    }
}

LogView.propTypes = {
    logs: PropTypes.array,
};

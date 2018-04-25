import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../../client';

export default class LogView extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            entries: [],
            streamRunning: false,
        };

        this.readLogs = this.readLogs.bind(this);
        this.scrollToBottom = this.scrollToBottom.bind(this);
    }

    scrollToBottom() {
        this.messagesEnd.scrollIntoView({ behavior: 'smooth' });
    }

    componentDidMount() {
        this.scrollToBottom();
        if (!this.state.streamRunning) this.readLogs().catch((err) => {
            this.setState({ streamRunning: false });
        });
    }

    componentDidUpdate() {
        this.scrollToBottom();
        if (!this.state.streamRunning) this.readLogs().catch((err) => {
            this.setState({ streamRunning: false });
        });
    }

    async readLogs() {
        const decoder = new TextDecoder('utf-8');
        let buffer = '';
        const stream = () => {
            if (!this.props.logs) {
                this.setState({
                    streamRunning: false,
                });
                return;
            }

            return this.props.logs.read().then((data) => {
                const chunk = decoder.decode(data.value);
                const parts = chunk.split('\n')
                    .filter((p) => p.length > 0);
                if (parts.length === 0) return;

                parts[0] = buffer + parts[0];
                buffer = '';
                if (!chunk.endsWith('\n')) {
                    buffer = parts.pop();
                }

                this.setState({
                    entries: this.state.entries.concat(parts),
                });

                return stream();
            });
        };
        return stream();
    }

    render() {
        let i = 0;
        const entries = this.state.entries.map((l) => {
            i++;
            return (<code key={i} >{l}<br /></code>);
        });

        return (
            <div>
                <div style={{
                    flex: 1,
                    bottom: 0,
                    left: 0,
                    position: 'absolute'
                }}>
                    {entries}
                </div>
                <div style={{ float: 'left', clear: 'both' }}
                    ref={(el) => { this.messagesEnd = el; }}>
                </div>
            </div>
        );
    }
}

LogView.propTypes = {
    reader: PropTypes.instanceOf(ReadableStream),
};

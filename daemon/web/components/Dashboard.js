import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../client';
import LogView from './metrics/LogView';

export default class Dashboard extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            errored: false,
            logEntries: [],
        };
        this.getLogs = this.getLogs.bind(this);
        this.getMessage = this.getMessage.bind(this);
    }

    async getLogs() {
        this.setState({ errored: false, logEntries: [] });
        try {
            let resp;
            if (!this.props.container) {
                resp = await this.props.client.getContainerLogs();
            } else {
                resp = await this.props.client.getContainerLogs(this.props.container);
            }
            if (resp.status !== 200) this.setState({
                errored: true, logEntries: [],
            });

            const reader = resp.body.getReader();
            const decoder = new TextDecoder('utf-8');
            let buffer = '';
            const stream = () => {
                return reader.read().then((data) => {
                    const chunk = decoder.decode(data.value);
                    const parts = chunk.split('\n');

                    parts[0] = buffer + parts[0];
                    buffer = '';
                    if (!chunk.endsWith('\n')) {
                        buffer = parts.pop();
                    }

                    this.setState({
                        logEntries: this.state.logEntries.concat(parts),
                    });

                    return stream();
                });
            };
            stream();
        } catch (e) {
            this.setState({
                errored: true,
                logEntries: [],
            });
            console.error(e);
        }
    }

    componentDidMount() {
        this.getLogs();
    }

    componentDidUpdate(prevProps) {
        if (prevProps.container != this.props.container) {
            this.getLogs();
        }
    }

    getMessage() {
        if (this.state.errored) {
            return <p style={styles.underConstruction}>Yikes, something went wrong</p>;
        } else if (this.state.logEntries.length == 0) {
            return <p style={styles.underConstruction}>No logs to show</p>;
        }
    }

    render() {
        return (
            <div style={{
                display: 'flex',
                height: '100%',
                alignItems: 'center',
                justifyContent: 'center',
                position: 'relative'
            }}>
                {this.getMessage()}
                <LogView logs={this.state.logEntries} />
            </div>
        );
    }
}

Dashboard.propTypes = {
    container: PropTypes.string,
    client: PropTypes.instanceOf(InertiaClient),
};

const styles = {
    underConstruction: {
        textAlign: 'center',
        fontSize: 24,
        color: '#9f9f9f',
    }
};

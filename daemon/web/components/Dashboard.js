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
        try {
            let resp;
            if (this.props.container) {
                resp = await this.props.client.getContainerLogs();
            } else {
                resp = await this.props.client.getContainerLogs(this.props.container);
            }
            if (resp.status != 200) this.setState({ errored: true, logEntries: [] });
            const logs = await resp.json();
            console.log(logs)
        } catch (e) {
            console.log(e);
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
        this.getLogs();
        return (
            <div style={{ display: 'flex', height: '100%', alignItems: 'center', justifyContent: 'center' }}>
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
        color: '#9f9f9f'
    }
};

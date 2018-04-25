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
            switch: true,
            reader: null,
        };
        this.getLogs = this.getLogs.bind(this);
        this.getMessage = this.getMessage.bind(this);
    }

    async getLogs() {
        let resp;
        if (!this.props.container) {
            resp = await this.props.client.getContainerLogs();
        } else {
            resp = await this.props.client.getContainerLogs(this.props.container);
        }

        if (resp.status !== 200) Promise.reject(new Error('non-200 response'));

        const reader = resp.body.getReader();
        this.setState({ reader: reader });
    }

    componentDidMount() {
        this.getLogs().catch((err) => {
            this.setState({
                errored: true,
                reader: null,
            });
        });
    }

    componentDidUpdate(prevProps, prevState) {
        if (prevProps.container !== this.props.container) {
            this.setState({
                errored: false,
                reader: null,
            });
            this.getLogs().catch((err) => {
                this.setState({
                    errored: true,
                    reader: null,
                });
            });
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
                <LogView logs={this.state.reader} />
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

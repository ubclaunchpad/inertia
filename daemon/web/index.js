import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/App';
import InertiaClient from './client';
import './index.sass';

// Define where the Inertia daemon is hosted.
const daemonAddress = (process.env.NODE_ENV === 'development') ? '127.0.0.1:4303' : window.location.host;
const client = new InertiaClient(daemonAddress);
ReactDOM.render(
    <App client={client} />,
    document.getElementById('app')
);

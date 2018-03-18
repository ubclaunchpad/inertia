import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../client';

export default class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div>
        <p align="center">
          <img src="https://github.com/ubclaunchpad/inertia/blob/master/.static/inertia-with-name.png?raw=true"
            width="10%" />
        </p>
        <p align="center">
          This is the Inertia web client!
        </p>
      </div>
    );
  }
}

App.propTypes = {
	client: PropTypes.instanceOf(InertiaClient)
}

const styles = {
  container: {
    display: 'flex',
  },
};

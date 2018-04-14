import React from 'react';
import PropTypes from 'prop-types';

export default class Home extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div>
        <h1>Welcome to the Inertia dashboard!</h1>
      </div>
    );
  }
}

App.propTypes = {
  client: PropTypes.instanceOf(InertiaClient)
};

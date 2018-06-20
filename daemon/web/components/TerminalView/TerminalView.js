import React from 'react';
// import PropTypes from 'prop-types';
import styles from './index.sass';

export default class TerminalViewView extends React.Component {
  constructor(props) {
    // Expect logs to be list of string
    super(props);
  }

  render() {
      return (
          <div>
            <textarea readOnly value={this.props.logs.reduce((accumulator, currentVal) => {return accumulator + "\r\n" + currentVal;})}></textarea>
          </div>
    );
  }
}

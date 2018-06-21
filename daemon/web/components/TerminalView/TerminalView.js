import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const styles = {
  textarea: {
    position: 'relative',
    top: '1px',
    width: '100%',
    height: '70px',
    color: '#ffffff',
    backgroundColor: '#212b36',
    fontSize: '10px',
    resize: 'none',
  },
};

const TerminalViewView = props =>
  (
    <div>
      <textarea
        style={styles.textarea}
        readOnly
        value={props.logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
    </div>
  );


TerminalViewView.propTypes = {
  logs: PropTypes.string,
};

export default TerminalViewView;

// AppRegistry.registerComponent('TerminalViewView', () => App);


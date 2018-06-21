import React from 'react';
import PropTypes from 'prop-types';

const styles = {
  textArea: {
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

const TerminalView = props =>
  (
    <div>
      <textarea
        style={styles.textArea}
        readOnly
        value={props.logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
    </div>
  );


TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

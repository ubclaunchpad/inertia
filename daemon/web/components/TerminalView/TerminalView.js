import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const TerminalView = ({ logs }) => (
  <div className="terminalView">
    <textarea
      className="textArea"
      readOnly
      value={!logs ? '' : logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
  </div>
);


TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

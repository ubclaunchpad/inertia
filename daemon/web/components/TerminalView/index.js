import React from 'react';
import PropTypes from 'prop-types';

const TerminalView = ({ logs }) => (
  <div className="terminal-view">
    <textarea
      className="text-area"
      readOnly
      value={!logs ? '' : logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
  </div>
);

TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

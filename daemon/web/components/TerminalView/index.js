import React from 'react';
import PropTypes from 'prop-types';

import './index.sass';

const TerminalView = ({ logs }) => (
  <div className="terminal-view height-xxxl flex">
    <textarea
      className="text-area pos-relative pad-sides-s pad-ends-xxs fill-height fill-width bg-terminal color-white"
      readOnly
      value={!logs ? '' : logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
  </div>
);

TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

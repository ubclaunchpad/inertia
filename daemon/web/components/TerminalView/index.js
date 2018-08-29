import React from 'react';
import PropTypes from 'prop-types';

import './index.sass';

const TerminalView = ({ logs }) => (
  <div className="terminal-view fill-width height-xxxl">
    <textarea
      className="text-area pad-sides-s pad-ends-xxs fill-height fill-width bg-terminal"
      readOnly
      value={!logs ? '' : logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
  </div>
);

TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

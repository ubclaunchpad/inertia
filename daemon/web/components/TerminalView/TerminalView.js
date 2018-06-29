import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

const TerminalView = props =>
  (
    <div className="terminalView">
      <textarea
        className="textArea"
        readOnly
        value={!props.logs ? '' : props.logs.reduce((accumulator, currentVal) => accumulator + '\r\n' + currentVal)} />
    </div>
  );


TerminalView.propTypes = {
  logs: PropTypes.array,
};

export default TerminalView;

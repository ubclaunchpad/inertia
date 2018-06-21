import React from 'react';
import './index.sass';

const styles = {
  textarea: {
    position: 'relative',
    top: '1px',
    width:'100%',
    height: '70px',
    color: '#ffffff',
    backgroundColor:'#212b36',
    fontSize: '10px',
    resize: 'none',
  },
};

export default class TerminalViewView extends React.Component {
    constructor(props) {
        // Expects logs to be list of string
        super(props);
    }

    render() {
      const results = this.props.logs.reduce((accumulator, currentVal) => {return accumulator + "\r\n" + currentVal;})
      return (
        <div>
          <textarea
          style={styles.textarea}
          readOnly
          value = {results}>
          </textarea>
        </div>
      );
    }
}

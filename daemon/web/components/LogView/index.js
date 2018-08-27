import React from 'react';
import PropTypes from 'prop-types';

const styles = {
  innerContainer: {
    flex: 1,
    bottom: 0,
    left: 0,
    position: 'absolute',
  },
  logsContainer: {
    float: 'left',
    clear: 'both',
  },
};

export default class LogView extends React.Component {
  constructor(props) {
    super(props);

    this.getEntries = this.getEntries.bind(this);
    this.scrollToBottom = this.scrollToBottom.bind(this);
  }

  componentDidMount() {
    this.scrollToBottom();
  }

  componentDidUpdate() {
    this.scrollToBottom();
  }

  getEntries() {
    const { logs } = this.props;
    let i = 0;
    return logs.map((l) => {
      i += 1;
      return (
        <code key={i}>
          {l}
          <br />
        </code>
      );
    });
  }

  scrollToBottom() {
    this.messagesEnd.scrollIntoView({ behavior: 'smooth' });
  }

  render() {
    const resultList = this.getEntries();
    return (
      <div>
        <div style={styles.innerContainer}>
          {resultList}
        </div>

        <div
          style={styles.messageContainer}
          ref={(el) => { this.messagesEnd = el; }}
        />
      </div>
    );
  }
}

LogView.propTypes = {
  logs: PropTypes.arrayOf(PropTypes.string),
};

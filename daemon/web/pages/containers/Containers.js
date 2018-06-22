import React from 'react';
import { connect } from 'react-redux';

const styles = {
};

// const mocklogs = [
//   'log1asdasdasdasdasdasdasdssdasdasdssdasdasdssdasdasdssdasdasdsa',
//   'log2asdasdasdasdsdassdasdasdssdasdasdssdasdasdssdasdasdsdsdasds',
//   'log3dasdsdazxcxzsdasdasdssdasdasdssdasdasdssdasdasdsxxxxxxxxxx',
//   'log4dasdsdasdsdasdasdssdasdasdssdasdasdssdasdasdsxzczxczxs',
//   'log5dasdsdaasdsdasdasdssdasdasdssdasdasdssdasdasdsasdasdsds',
//   'log6dasdsdaszsdasdasdssdasdasdssdasdasdssdasdasdsxczxczxczxcwqdqds',
//   'log7dasdsdaxcsdasdasdssdasdasdssdasdasdssdasdasdszxczzxcsds',
// ];

class ContainersWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  render() {
    return (
      <div style={styles.container}>
        <h1>Hello!</h1>
      </div>
    );
  }
}
ContainersWrapper.propTypes = {};

const mapStateToProps = () => { return {}; };

const mapDispatchToProps = () => { return {}; };

const Containers = connect(mapStateToProps, mapDispatchToProps)(ContainersWrapper);

export default Containers;

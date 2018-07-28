import React from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import TerminalView from '../../components/TerminalView/TerminalView';
import IconHeader from '../../components/IconHeader/IconHeader';
import Status from '../../components/Status/Status';
import '../../pages/containers/index.sass';

const mocklogs = [
  'log1asdasdasdasdasdasdasdssdasdasdssdasdasdssdasdasdssdasdasdsa',
  'log2asdasdasdasdsdassdasdasdssdasdasdssdasdasdssdasdasdsdsdasds',
  'log3dasdsdazxcxzsdasdasdssdasdasdssdasdasdssdasdasdsxxxxxxxxxx',
  'log4dasdsdasdsdasdasdssdasdasdssdasdasdssdasdasdsxzczxczxs',
  'log5dasdsdaasdsdasdasdssdasdasdssdasdasdssdasdasdsasdasdsds',
  'log6dasdsdaszsdasdasdssdasdasdssdasdasdssdasdasdsxczxczxczxcwqdqds',
  'log7dasdsdaxcsdasdasdssdasdasdssdasdasdssdasdasdszxczzxcsds',
];

class ContainersWrapper extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  render() {
    return (
      <div>
        <IconHeader type="containers" title="/inertia-deploy-test_dev_1" />
        <div className="containerInfo" >
          <Status title="Status:" status="Active" />
          <h3>Last Updated:</h3>
          <h4>{this.props.dateUpdated}</h4>
        </div>
        <TerminalView logs={mocklogs} />
      </div>
    );
  }
}
ContainersWrapper.propTypes = {
  dateUpdated: PropTypes.string,
};

const mapStateToProps = () => { return {}; };

const mapDispatchToProps = () => { return {}; };

const Containers = connect(mapStateToProps, mapDispatchToProps)(ContainersWrapper);

export default Containers;

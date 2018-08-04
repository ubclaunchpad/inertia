import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

export const TableCell = ({ children, style }) => (
  <div className="table-td" style={style}>
    {children}
  </div>
);
TableCell.propTypes = {
  style: PropTypes.object,
  children: PropTypes.any,
};

export const TableRow = ({ style, children }) => (
  <div className="table-tr" style={style}>
    {children}
  </div>
);
TableRow.propTypes = {
  style: PropTypes.object,
  children: PropTypes.any,
};

export class TableRowExpandable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      expanded: false,
    };
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(expanded) {
    this.setState({ expanded: !expanded });
  }

  render() {
    const {
      style,
      children,
      panel,
      onClick,
    } = this.props;

    return (
      <div
        className="table-tr-expandable"
        onClick={onClick}
        style={style}>
        <div
          className={`table-tr-expandable-inner ${this.state.expanded && 'expanded'}`}
          onClick={() => this.handleClick(this.state.expanded)}>
          {children}
        </div>
        <div
          style={{ height: this.state.expanded ? this.props.height : 0 }}
          className={`table-tr-panel ${this.state.expanded && 'expanded'}`}>
          {panel}
        </div>
      </div>
    );
  }
}
TableRowExpandable.propTypes = {
  style: PropTypes.object,
  children: PropTypes.arrayOf(TableCell),
  panel: PropTypes.any,
  height: PropTypes.number,
  onClick: PropTypes.func,
};

export const TableHeader = ({ children, style }) => (
  <div className="table-thead" style={style}>
    {children}
  </div>
);
TableHeader.propTypes = {
  style: PropTypes.object,
  children: PropTypes.objectOf(TableRow),
};

export const TableBody = ({ children, style }) => (
  <div className="table-tbody" style={style}>
    {children}
  </div>
);
TableBody.propTypes = {
  style: PropTypes.object,
  children: PropTypes.arrayOf(TableRow),
};

export const Table = ({ children, style }) => (
  <div className="table" style={style}>
    {children}
  </div>
);
Table.propTypes = {
  style: PropTypes.object,
  children: PropTypes.any,
};

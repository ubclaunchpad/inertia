import React from 'react';
import PropTypes from 'prop-types';

export const Table = ({ children, className = '', style }) => (
  <div className={`shadow rounded pos-relative bg-white ${className}`} style={style}>
    {children}
  </div>
);
Table.propTypes = {
  children: PropTypes.any,
  className: PropTypes.string,
  style: PropTypes.object,
};

export const TableCell = ({ children, className = '', style }) => (
  <div className={`flex ai-center f-single height-xl ${className}`} style={style}>
    {children}
  </div>
);
TableCell.propTypes = {
  style: PropTypes.object,
  children: PropTypes.any,
  className: PropTypes.string,
};

export const TableRow = ({ children, className = '', style }) => (
  <div className={`flex fill-width pad-sides-l height-xl ${className}`} style={style}>
    {children}
  </div>
);
TableRow.propTypes = {
  style: PropTypes.object,
  children: PropTypes.any,
  className: PropTypes.string,
};

export const TableHeader = ({ children, className = '', style }) => (
  <div className={`border-end ${className}`} style={style}>
    {children}
  </div>
);
TableHeader.propTypes = {
  children: PropTypes.objectOf(TableRow),
  className: PropTypes.string,
  style: PropTypes.object,
};

export const TableBody = ({ children, className = '', style }) => (
  <div className={`table-body ${className}`} style={style}>
    {children}
  </div>
);
TableBody.propTypes = {
  children: PropTypes.arrayOf(TableRow),
  className: PropTypes.string,
  style: PropTypes.object,
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
      children,
      panel,
      onClick,
      height,
      style,
      className = '',
    } = this.props;
    const { expanded } = this.state;

    return (
      <div
        className={`flex wrap fill-width pos-relative ${className}`}
        onClick={onClick}
        style={style}>
        <div
          className={`flex fill-width pos-relative pad-sides-l clickable
            hover-bg-highlight-light ${expanded && 'bg-highlight-light '}`}
          onClick={() => this.handleClick(expanded)}>
          {children}
        </div>
        <div
          style={{ height: expanded ? height : 0 }}
          className={`transition-ease fill-width scroll
            ${expanded ? 'visible' : 'hidden'}`}>
          {panel}
        </div>
      </div>
    );
  }
}
TableRowExpandable.propTypes = {
  children: PropTypes.arrayOf(TableCell),
  panel: PropTypes.any,
  height: PropTypes.number,
  onClick: PropTypes.func,
  className: PropTypes.string,
  style: PropTypes.object,
};

import React from 'react';
import PropTypes from 'prop-types';

import footerLaunchpad from '../../assets/icons/launchpad-logo-lightblue.png';
import './index.sass';

const Footer = ({ version }) => {
  const icons = [
    {
      id: 1,
      link: 'https://github.com/ubclaunchpad/inertia',
      name: 'fab fa-github',
    },
    {
      id: 2,
      link: 'https://medium.com/ubc-launch-pad-software-engineering-blog',
      name: 'fab fa-medium-m',
    },
    {
      id: 3,
      link: 'https://github.com/ubclaunchpad/inertia/issues/new/choose',
      name: 'fas fa-comments',
    },
  ].map(socialFooterIcon => (
    <span className="icons">
      <a key={socialFooterIcon.id} href={socialFooterIcon.link}>
        <span className={socialFooterIcon.name} />
      </a>
    </span>));
  return (
    <footer className="footer">
      <span className="left-side"><p className="daemon-version">Daemon {version}</p></span>
      <span className="right-side">
        {icons}
        <a href="https://www.ubclaunchpad.com">
          <span className="launchpad"><img src={footerLaunchpad} height="18" width="130" alt="ubc launchpad" /></span>
        </a>
      </span>
    </footer>
  );
};

Footer.propTypes = {
  version: PropTypes.string,
};

export default Footer;

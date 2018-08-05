import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';
import FooterLaunchpad from '../../assets/logo/launchpad-logo-light-blue-01.svg';

const backgroundColor = '#474d5e';

const Footer = ({ version }) => {
  const icons = [
    {
      id: 1,
      link: 'https://github.com/ubclaunchpad/inertia',
      name: 'fab fa-github',
    },
    {
      id: 2,
      link: 'https://github.com/ubclaunchpad/inertia/issues/new/choose',
      name: 'fas fa-comments',
    },
    {
      id: 3,
      link: 'https://medium.com/ubc-launch-pad-software-engineering-blog',
      name: 'fab fa-medium-m',
    },
  ].map(socialFooterIcon => (
    <span className="icons">
      <a
        key={socialFooterIcon.id}
        href={socialFooterIcon.link}
        style={{ backgroundColor }}
        rel="noopener noreferrer"
        target="_blank">
        <span className={socialFooterIcon.name} />
      </a>
    </span>));
  return (
    <footer className="footer" style={{ backgroundColor }}>
      <span className="left-side">
        <p className="daemon-version">
inertiad
          {version}
        </p>
      </span>
      <span className="right-side">
        {icons}
        <span className="icons">
          <div>
            <a href="https://www.ubclaunchpad.com" rel="noopener noreferrer" target="_blank">
              <FooterLaunchpad className="launchpad" />
            </a>
          </div>
        </span>
      </span>
    </footer>
  );
};

Footer.propTypes = {
  version: PropTypes.string,
};

export default Footer;
